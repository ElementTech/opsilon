package kubengine

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	_ "unsafe"

	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/jatalocks/opsilon/internal/engine"
	"github.com/jatalocks/opsilon/internal/internaltypes"
	"github.com/jatalocks/opsilon/internal/logger"
	"golang.org/x/exp/slices"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	_ "k8s.io/kubectl/pkg/cmd/cp"
	"k8s.io/kubectl/pkg/scheme"
)

type Client struct {
	k8s    kubernetes.Interface
	ns     string
	config rest.Config
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
					v1.ResourceName(v1.ResourceStorage): resource.MustParse("1Gi"),
				},
			},
			VolumeMode: &fs,
		},
		Status: v1.PersistentVolumeClaimStatus{
			Phase:       v1.ClaimBound,
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): resource.MustParse("1Gi"),
			},
		},
	}

	api := c.k8s.CoreV1()
	claim, errGo := api.PersistentVolumeClaims(c.ns).Create(ctx, createOpts, metav1.CreateOptions{})
	logger.HandleErr(errGo)

	if mount {
		wd, err := os.Getwd()
		logger.HandleErr(err)

		VolumeMounts := []v1.VolumeMount{{Name: volumeName, MountPath: "/app"}}
		Volumes := []v1.Volume{{
			Name:         volumeName,
			VolumeSource: v1.VolumeSource{PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{ClaimName: claim.Name}},
		}}

		string_uuid := (uuid.New()).String()

		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: string_uuid},
			Spec: v1.PodSpec{
				RestartPolicy: v1.RestartPolicyNever,
				Volumes:       Volumes,
				Containers: []v1.Container{
					v1.Container{
						Name:         "keepalive",
						Image:        "busybox",
						Command:      []string{"/bin/sh"},
						Args:         []string{"-c", fmt.Sprintf("until ((ls /%s)); do echo sleeping; sleep 2; done;cp -r /%s/. ./;find / -path '*/%s/*' -delete;rm -rf /app/lost+found", path.Base(wd), path.Base(wd), path.Base(wd))},
						WorkingDir:   "/app",
						VolumeMounts: VolumeMounts,
					},
				},
			},
		}

		defer func() {
			recover()
			if err := c.DeletePod(ctx, string_uuid); err != nil {
				log.Printf("Error deleting pod: %v", err)
			}
		}()

		_, err = c.k8s.CoreV1().
			Pods(c.ns).
			Create(ctx, pod, metav1.CreateOptions{})
		logger.HandleErr(err)
		LwWhite := logger.NewLogWriter(func(str string, color color.Attribute) {
			logger.Custom(color, fmt.Sprintf("[%s] %s", "Mounting working directory to volume", str))
		}, color.FgWhite)

		err = c.waitPod(ctx, string_uuid, LwWhite, "Terminated") // terminated == running in this case. because mounter does not have init containers.
		logger.HandleErr(err)

		copyToPod(c, wd, "/", string_uuid, ctx)

		err = c.waitPod(ctx, string_uuid, LwWhite, "Mounted")
		logger.HandleErr(err)

	}

	return volumeName, claim
}

func (c *Client) RemoveVolume(ctx context.Context, vol string, claim *v1.PersistentVolumeClaim) {
	recover()
	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	err := c.k8s.CoreV1().PersistentVolumeClaims(c.ns).Delete(ctx, claim.Name, deleteOptions)
	fmt.Println(err)
	// err = c.k8s.CoreV1().PersistentVolumes().Delete(ctx, vol, deleteOptions)
	// fmt.Println(err)
}

func (c *Client) CreatePod(ctx context.Context, name string, image string, command []string, envs []internaltypes.Env, volume string, claim *v1.PersistentVolumeClaim) (error, v1.Pod) {
	envVar := ToV1Env(envs)
	VolumeMounts := []v1.VolumeMount{{Name: "emptyoutput", MountPath: "/output"}}
	Volumes := []v1.Volume{{
		Name:         "emptyoutput",
		VolumeSource: v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{Medium: v1.StorageMediumMemory}},
	}}
	if volume != "" {
		VolumeMounts = append(VolumeMounts, v1.VolumeMount{Name: volume, MountPath: "/app"})
		Volumes = append(Volumes, v1.Volume{
			Name:         volume,
			VolumeSource: v1.VolumeSource{PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{ClaimName: claim.Name}},
		})
	} else {
		VolumeMounts = append(VolumeMounts, v1.VolumeMount{Name: "emptydir", MountPath: "/app"})
		Volumes = append(Volumes, v1.Volume{
			Name:         "emptydir",
			VolumeSource: v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{Medium: v1.StorageMediumMemory}},
		})
	}

	// if volumeOutput != "" {
	// 	VolumeMounts = append(VolumeMounts, v1.VolumeMount{Name: volumeOutput, MountPath: "/output"})
	// 	Volumes = append(Volumes, v1.Volume{
	// 		Name:         volumeOutput,
	// 		VolumeSource: v1.VolumeSource{PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{ClaimName: claimOutput.Name}},
	// 	})
	// }
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: v1.PodSpec{
			RestartPolicy: v1.RestartPolicyNever,
			Volumes:       Volumes,
			InitContainers: []v1.Container{
				v1.Container{
					Name:         "main",
					Image:        image,
					Command:      command,
					WorkingDir:   "/app",
					Env:          *envVar,
					VolumeMounts: VolumeMounts,
				},
			},
			Containers: []v1.Container{
				v1.Container{
					Name:         "keepalive",
					Image:        "busybox",
					WorkingDir:   "/app",
					Command:      []string{"/bin/sh", "-c", "sleep 60"},
					VolumeMounts: VolumeMounts,
				},
			},
		},
	}

	_, err := c.k8s.CoreV1().
		Pods(c.ns).
		Create(ctx, pod, metav1.CreateOptions{})

	return err, *pod
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
		if status := p.Status.InitContainerStatuses[0].State.Terminated; status != nil {
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
	recover()
	deletePolicy := metav1.DeletePropagationBackground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	return c.k8s.CoreV1().
		Pods(c.ns).
		Delete(ctx, name, deleteOptions)
}

// func (c *Client) DeleteNamespace(ctx context.Context) {
// 	deletePolicy := metav1.DeletePropagationForeground
// 	deleteOptions := metav1.DeleteOptions{
// 		PropagationPolicy: &deletePolicy,
// 	}
// 	err := c.k8s.CoreV1().Namespaces().Delete(ctx, c.ns, deleteOptions)
// 	logger.HandleErr(err)
// }

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
	clientCfg, _ := clientcmd.NewDefaultClientConfigLoadingRules().Load()
	namespace := ""
	if len(clientCfg.Contexts) != 0 {
		namespace = clientCfg.Contexts[clientCfg.CurrentContext].Namespace
	}

	if namespace == "" {
		namespace = "default"
	}
	return &Client{k8s: k8s, ns: namespace, config: *config}, nil
}

func toPodName(stage internaltypes.Stage) string {
	return fmt.Sprint(strings.ReplaceAll(clearString(fmt.Sprint(stage.Stage+"-"+stage.ID)), " ", "-") + "-" + (uuid.New()).String())
}

func (cli *Client) KubeEngine(wg *sync.WaitGroup, sID string, ctx context.Context, w internaltypes.Workflow, vol string, claim *v1.PersistentVolumeClaim, allOutputs map[string][]internaltypes.Env, skippedStages *[]string, results chan internaltypes.Result) {
	defer wg.Done()
	idx := slices.IndexFunc(w.Stages, func(c internaltypes.Stage) bool { return c.ID == sID })
	stage := w.Stages[idx]
	result := internaltypes.Result{Stage: stage}
	// volOutput, claimOutput := cli.CreateVolume(ctx, false)
	// defer cli.RemoveVolume(ctx, volOutput, claimOutput)

	allEnvs, needSplit, LwWhite, LwCrossed := engine.PrepareStage(w.Env, stage.Env, w.Input, stage.Needs, allOutputs, stage.Stage, stage.ID, &result)

	if !engine.EvaluateCondition(stage.If, allEnvs, LwWhite) {
		*skippedStages = append(*skippedStages, stage.ID)
		result.Skipped = true
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
			result.Skipped = true
			LwCrossed.Println("Stage Skipped due to needed stage skipped")
		} else {
			if stage.Clean {
				// volClean, claimClean := cli.CreateVolume(ctx, false)
				success := cli.RunStageKubernetes(stage, ctx, allEnvs, w.Image, "", nil, LwWhite, &result, allOutputs)
				result.Result = success
				// defer cli.RemoveVolume(ctx, volClean, claimClean)
			} else {
				success := cli.RunStageKubernetes(stage, ctx, allEnvs, w.Image, vol, claim, LwWhite, &result, allOutputs)
				result.Result = success
			}
		}

	}
	results <- result
}

func ToV1Env(envs []internaltypes.Env) *[]v1.EnvVar {
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

func (c *Client) createPodWatcher(ctx context.Context, resName string) (watch.Interface, error) {
	fieldSelector := fmt.Sprintf("metadata.name=%s", resName)

	opts := metav1.ListOptions{
		TypeMeta:      metav1.TypeMeta{},
		FieldSelector: fieldSelector,
	}

	return c.k8s.CoreV1().Pods(c.ns).Watch(ctx, opts)
}

func (c *Client) waitPod(ctx context.Context, resName string, LwWhite *logger.MyLogWriter, state string) error {
	watcher, err := c.createPodWatcher(ctx, resName)
	if err != nil {
		return err
	}

	defer watcher.Stop()

	for {
		select {
		case event := <-watcher.ResultChan():
			pod := event.Object.(*v1.Pod)
			if state == "Running" {
				if len(pod.Status.InitContainerStatuses) > 0 {
					if pod.Status.InitContainerStatuses[0].State.Waiting == nil {
						return nil
					}
				}

			} else if state == "Terminated" {
				if pod.Status.Phase == v1.PodRunning {
					return nil
				}
				if pod.Status.Phase == v1.PodFailed {
					return nil
				}
			} else if state == "Mounted" {
				if pod.Status.Phase == v1.PodSucceeded {
					return nil
				}
				if pod.Status.Phase == v1.PodFailed {
					return errors.New("mounting failed")
				}
			}
			// LwWhite.Write([]byte(fmt.Sprintf("The POD \"%s\" is running/success\n", resName)))

		case <-ctx.Done():
			// LwWhite.Write([]byte(fmt.Sprintf("Exit from waitPodRunning for POD \"%s\" because the context is done", resName)))
			return nil
		}
	}
}

func (c *Client) getPodLogs(ctx context.Context, podName string, LwWhite *logger.MyLogWriter) error {
	count := int64(100)
	podLogOptions := v1.PodLogOptions{
		Follow:    true,
		TailLines: &count,
		Container: "main",
	}

	err := c.waitPod(ctx, podName, LwWhite, "Running")
	logger.HandleErr(err)

	podLogRequest := c.k8s.CoreV1().
		Pods(c.ns).
		GetLogs(podName, &podLogOptions)
	stream, err := podLogRequest.Stream(ctx)
	if err != nil {
		return err
	}
	defer stream.Close()
	for {
		buf := make([]byte, 2000)
		numBytes, err := stream.Read(buf)
		if err == io.EOF {
			break
		}
		if numBytes == 0 {
			continue
		}
		if err != nil {
			return err
		}
		message := string(buf[:numBytes])
		LwWhite.Write([]byte(message))
	}
	return nil
}

func (cli *Client) RunStageKubernetes(s internaltypes.Stage, ctx context.Context, envs []internaltypes.Env, globalImage string, volume string, claim *v1.PersistentVolumeClaim, LwWhite *logger.MyLogWriter, result *internaltypes.Result, allOutputs map[string][]internaltypes.Env) bool {
	LwWhite.Write([]byte(fmt.Sprintf("Running Stage with the following variables: %s\n", engine.GenEnv(envs))))
	envs = append(envs, []internaltypes.Env{{Name: "OUTPUT", Value: "/output/output"}}...)
	if s.Image != "" {
		globalImage = s.Image
	}

	podName := toPodName(s)
	err, _ := cli.CreatePod(
		ctx,
		podName,
		globalImage,
		s.Script,
		envs,
		volume,
		claim,
	)
	logger.HandleErr(err)

	err = cli.getPodLogs(ctx, podName, LwWhite)
	logger.HandleErr(err)

	err = cli.waitPod(ctx, podName, LwWhite, "Terminated")
	logger.HandleErr(err)

	dirArt, err := os.MkdirTemp("", "artifacts")
	logger.HandleErr(err)
	// defer os.RemoveAll(dirArt)

	// err = cli.waitPod(ctx, podName, LwWhite, v1.PodRunning)
	// logger.HandleErr(err)

	if len(s.Artifacts) > 0 {
		err = copyFromPod(cli, "/app", dirArt, podName)
		logger.HandleErr(err)
		engine.ExtractArtifacts(path.Join(dirArt, "app"), s)
	}
	err = copyFromPod(cli, "/output", dirArt, podName)
	if err != nil {
		logger.Error(err.Error())
	}
	outputMap, err := engine.ReadPropertiesFile(path.Join(dirArt, "/output/output"))
	if err != nil {
		logger.Error(err.Error())
	}
	allOutputs[s.ID] = outputMap
	result.Outputs = outputMap

	// err = cli.waitPod(ctx, podName, LwWhite, v1.PodRunning)
	// logger.HandleErr(err)

	recover()

	exitCode, err := cli.GetPodExitCode(ctx, podName)
	logger.HandleErr(err)

	if err := cli.DeletePod(ctx, podName); err != nil {
		log.Printf("Error deleting pod: %v", err)
	}

	return exitCode == 0

	// logger.Info(fmt.Sprint(exitCode))
}
func getPrefix(file string) string {
	return strings.TrimLeft(file, "/")
}

//go:linkname cpMakeTar k8s.io/kubectl/pkg/cmd/cp.makeTar
func cpMakeTar(srcPath, destPath string, writer io.Writer) error

func copyToPod(cli *Client, srcPath string, destPath string, podName string, ctx context.Context) error {
	restconfig := cli.config

	reader, writer := io.Pipe()
	if destPath != "/" && strings.HasSuffix(string(destPath[len(destPath)-1]), "/") {
		destPath = destPath[:len(destPath)-1]
	}
	destPath = destPath + "/" + path.Base(srcPath)
	go func() {
		defer writer.Close()
		err := cpMakeTar(srcPath, destPath, writer)
		if err != nil {
			fmt.Println(err.Error())
		}
	}()
	cmdArr := []string{"tar", "-xf", "-"}
	destDir := path.Dir(destPath)
	if len(destDir) > 0 {
		cmdArr = append(cmdArr, "-C", destDir)
	}
	//remote shell.
	req := cli.k8s.CoreV1().RESTClient().
		Post().
		Namespace(cli.ns).
		Resource("pods").
		Name(podName).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Container: "keepalive",
			Command:   cmdArr,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(&restconfig, "POST", req.URL())
	if err != nil {
		log.Fatalf("error %s\n", err)
		return err
	}
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  reader,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Tty:    false,
	})
	if err != nil {
		log.Fatalf("error %s\n", err)
		return err
	}
	return nil
}

func copyFromPod(cli *Client, srcPath string, destPath string, podName string) error {
	restconfig := cli.config
	reader, outStream := io.Pipe()
	//todo some containers failed : tar: Refusing to write archive contents to terminal (missing -f option?) when execute `tar cf -` in container
	cmdArr := []string{"tar", "cf", "-", srcPath}
	req := cli.k8s.CoreV1().RESTClient().Get().
		Namespace(cli.ns).
		Resource("pods").
		Name(podName).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Container: "keepalive",
			Command:   cmdArr,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(&restconfig, "POST", req.URL())
	if err != nil {
		return err
	}
	go func() {
		defer outStream.Close()
		err = exec.Stream(remotecommand.StreamOptions{
			Stdin:  os.Stdin,
			Stdout: outStream,
			Stderr: os.Stderr,
			Tty:    false,
		})
		fmt.Printf("error %s\n", err)
	}()
	prefix := getPrefix(srcPath)
	prefix = path.Clean(prefix)
	prefix = cpStripPathShortcuts(prefix)
	destPath = path.Join(destPath, path.Base(prefix))
	err = untarAll(reader, destPath, prefix)
	return err
}

//go:linkname cpStripPathShortcuts k8s.io/kubectl/pkg/cmd/cp.stripPathShortcuts
func cpStripPathShortcuts(p string) string

func untarAll(reader io.Reader, destDir, prefix string) error {
	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}

		if !strings.HasPrefix(header.Name, prefix) {
			return fmt.Errorf("tar contents corrupted")
		}

		mode := header.FileInfo().Mode()
		destFileName := filepath.Join(destDir, header.Name[len(prefix):])

		baseName := filepath.Dir(destFileName)
		if err := os.MkdirAll(baseName, 0755); err != nil {
			return err
		}
		if header.FileInfo().IsDir() {
			if err := os.MkdirAll(destFileName, 0755); err != nil {
				return err
			}
			continue
		}

		evaledPath, err := filepath.EvalSymlinks(baseName)
		if err != nil {
			return err
		}

		if mode&os.ModeSymlink != 0 {
			linkname := header.Linkname

			if !filepath.IsAbs(linkname) {
				_ = filepath.Join(evaledPath, linkname)
			}

			if err := os.Symlink(linkname, destFileName); err != nil {
				return err
			}
		} else {
			outFile, err := os.Create(destFileName)
			if err != nil {
				return err
			}
			defer outFile.Close()
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return err
			}
			if err := outFile.Close(); err != nil {
				return err
			}
		}
	}

	return nil
}
