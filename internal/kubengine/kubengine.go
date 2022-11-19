package kubengine

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jatalocks/opsilon/internal/engine"
	"github.com/jatalocks/opsilon/internal/logger"
	"golang.org/x/exp/slices"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	k8s kubernetes.Interface
	ns  string
}

type RequestBody struct {
	Image   string   `json:"image"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

type ResponseBody struct {
	ExitCode int32  `json:"exit_code"`
	Output   string `json:"output"`
}

func (c *Client) CreateVolume(ctx context.Context, mount bool) (string, *v1.PersistentVolumeClaim) {
	volume, errGo := uuid.NewRandom()
	logger.HandleErr(errGo)
	volumeName := volume.String()

	fs := v1.PersistentVolumeFilesystem
	createOpts := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: volumeName,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceName(v1.ResourceStorage): resource.MustParse("10Gi"),
				},
			},
			VolumeMode: &fs,
		},
		Status: v1.PersistentVolumeClaimStatus{
			Phase:       v1.ClaimBound,
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): resource.MustParse("10Gi"),
			},
		},
	}

	api := c.k8s.CoreV1()
	claim, errGo := api.PersistentVolumeClaims(c.ns).Create(ctx, createOpts, metav1.CreateOptions{})
	logger.HandleErr(errGo)

	return volumeName, claim
}

func (c *Client) RemoveVolume(ctx context.Context, vol string, claim *v1.PersistentVolumeClaim) {
	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	c.k8s.CoreV1().PersistentVolumeClaims(c.ns).Delete(ctx, claim.Name, deleteOptions)
	c.k8s.CoreV1().PersistentVolumes().Delete(ctx, vol, deleteOptions)
}

func (c *Client) CreatePod(ctx context.Context, name string, image string, command []string, envs []engine.Env, volume string, claim *v1.PersistentVolumeClaim, volumeOutput string, claimOutput *v1.PersistentVolumeClaim) error {
	envVar := ToV1Env(envs)
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: v1.PodSpec{
			RestartPolicy: v1.RestartPolicyNever,
			Volumes: []v1.Volume{{
				Name:         volume,
				VolumeSource: v1.VolumeSource{PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{ClaimName: claim.Name}},
			}, {
				Name:         volumeOutput,
				VolumeSource: v1.VolumeSource{PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{ClaimName: claimOutput.Name}},
			}},
			Containers: []v1.Container{
				v1.Container{
					Name:         "main",
					Image:        image,
					Command:      command,
					WorkingDir:   "/app",
					Env:          *envVar,
					VolumeMounts: []v1.VolumeMount{{Name: volume, MountPath: "/app"}, {Name: volumeOutput, MountPath: "/output", SubPath: "output"}},
				},
			},
		},
	}

	_, err := c.k8s.CoreV1().
		Pods(c.ns).
		Create(ctx, pod, metav1.CreateOptions{})

	return err
}

func (c *Client) GetPodExitCode(ctx context.Context, name string) (int32, error) {
	var exitCode int32
	podCli := c.k8s.CoreV1().Pods(c.ns)
	err := wait.PollImmediate(3*time.Second, 2*time.Minute, func() (bool, error) {
		p, err := podCli.Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if len(p.Status.ContainerStatuses) == 0 {
			return false, nil
		}
		if status := p.Status.ContainerStatuses[0].State.Terminated; status != nil {
			exitCode = status.ExitCode
			return true, nil
		}
		return false, nil
	})
	return exitCode, err
}

func (c *Client) GetPodStdOut(ctx context.Context, name string) (string, error) {
	stdout, err := c.k8s.CoreV1().
		Pods(c.ns).
		GetLogs(name, &v1.PodLogOptions{}).
		Do(ctx).
		Raw()

	if err != nil {
		return "", err
	}
	return string(stdout), nil
}

func (c *Client) DeletePod(ctx context.Context, name string) error {
	return c.k8s.CoreV1().
		Pods(c.ns).
		Delete(ctx, name, metav1.DeleteOptions{})
}

func (c *Client) DeleteNamespace(ctx context.Context) {
	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	err := c.k8s.CoreV1().Namespaces().Delete(ctx, c.ns, deleteOptions)
	logger.HandleErr(err)
}

func NewClient() (*Client, error) {

	config, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig :=
			clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}

	}
	k8s, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	string_uuid := (uuid.New()).String()
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: string_uuid,
		},
	}
	k8s.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})
	return &Client{k8s: k8s, ns: ns.Name}, nil
}
func (cli *Client) KubeEngine(wg *sync.WaitGroup, sID string, ctx context.Context, w engine.Workflow, vol string, claim *v1.PersistentVolumeClaim, allOutputs map[string][]engine.Env, skippedStages *[]string) {
	defer wg.Done()

	idx := slices.IndexFunc(w.Stages, func(c engine.Stage) bool { return c.ID == sID })
	stage := w.Stages[idx]

	volOutput, claimOutput := cli.CreateVolume(ctx, false)
	defer cli.RemoveVolume(ctx, volOutput, claimOutput)

	allEnvs, needSplit, LwWhite, LwCrossed := engine.PrepareStage(w.Env, stage.Env, w.Input, stage.Needs, allOutputs, stage.Stage, stage.ID)

	if !engine.EvaluateCondition(stage.If, allEnvs, LwWhite) {
		*skippedStages = append(*skippedStages, stage.ID)
		LwCrossed.Println("Stage Skipped due to IF condition")
	} else {
		toSkip := false
		for _, skipped := range *skippedStages {
			for _, need := range needSplit {
				if need == skipped {
					toSkip = true
				}
			}
		}
		if toSkip {
			*skippedStages = append(*skippedStages, stage.ID)
			LwCrossed.Println("Stage Skipped due to needed stage skipped")
		} else {
			if stage.Clean {
				volClean, claimClean := cli.CreateVolume(ctx, false)
				cli.RunStageKubernetes(stage, ctx, allEnvs, w.Image, volClean, claimClean, volOutput, claimOutput, LwWhite)
			} else {
				cli.RunStageKubernetes(stage, ctx, allEnvs, w.Image, vol, claim, volOutput, claimOutput, LwWhite)
			}
		}

	}

	// outputMap, err := ReadPropertiesFile(outputPath)
	// logger.HandleErr(err)
	// allOutputs[stage.ID] = outputMap

	// err := cli.CreatePod(
	// 	ctx,
	// 	"rtw",
	// 	"python:3.8",
	// 	"python",
	// 	[]string{"-c", "print(\"hello world\")"},
	// )
	// logger.HandleErr(err)
	// exitCode, err := cli.GetPodExitCode(ctx, "rtw")
	// logger.HandleErr(err)
	// logger.HandleErr(err)
	// stdout, err := cli.GetPodStdOut(ctx, "rtw")
	// logger.HandleErr(err)
	// go func() {
	// 	if err := cli.DeletePod(ctx, "rtw"); err != nil {
	// 		log.Printf("Error deleting pod: %v", err)
	// 	}
	// }()
	// logger.Info(fmt.Sprint(exitCode))
	// logger.Free(stdout)
}

func ToV1Env(envs []engine.Env) *[]v1.EnvVar {
	envVar := []v1.EnvVar{}
	for _, v := range envs {
		envVar = append(envVar, v1.EnvVar{Name: v.Name, Value: v.Value})
	}
	return &envVar
}

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

func clearString(str string) string {
	return nonAlphanumericRegex.ReplaceAllString(str, "")
}

func (cli *Client) RunStageKubernetes(s engine.Stage, ctx context.Context, envs []engine.Env, globalImage string, volume string, claim *v1.PersistentVolumeClaim, volumeOutput string, claimOutput *v1.PersistentVolumeClaim, LwWhite *logger.MyLogWriter) {
	LwWhite.Write([]byte(fmt.Sprintf("Running Stage with the following variables: %s\n", engine.GenEnv(envs))))
	envs = append(envs, []engine.Env{{Name: "OUTPUT", Value: "/output/output"}}...)
	if s.Image != "" {
		globalImage = s.Image
	}

	// resp, err := cli.ContainerCreate(ctx, &container.Config{
	// 	Image:      globalImage,
	// 	Env:        allEnvs,
	// 	Cmd:        s.Script,
	// 	WorkingDir: "/app",
	// 	Tty:        false,
	// }, &hostConfig, nil, nil, "")
	podName := strings.ReplaceAll(clearString(fmt.Sprint(s.Stage+"-"+s.ID)), " ", "-")
	err := cli.CreatePod(
		ctx,
		podName,
		globalImage,
		s.Script,
		envs,
		volume,
		claim,
		volumeOutput,
		claimOutput,
	)
	logger.HandleErr(err)
	exitCode, err := cli.GetPodExitCode(ctx, podName)
	logger.HandleErr(err)
	logger.HandleErr(err)
	stdout, err := cli.GetPodStdOut(ctx, podName)
	logger.HandleErr(err)
	go func() {
		if err := cli.DeletePod(ctx, podName); err != nil {
			log.Printf("Error deleting pod: %v", err)
		}
	}()
	logger.Info(fmt.Sprint(exitCode))
	logger.Free(stdout)
}
