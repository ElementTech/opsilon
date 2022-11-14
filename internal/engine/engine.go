package engine

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/Knetic/govaluate"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/fatih/color"
	"github.com/jatalocks/opsilon/internal/logger"
	cp "github.com/otiai10/copy"
	"golang.org/x/exp/slices"
)

type Input struct {
	Name     string `mapstructure:"name" validate:"nonzero,nowhitespace"`
	Default  string `mapstructure:"default"`
	Optional bool   `mapstructure:"optional,omitempty"`
}

type Stage struct {
	Stage     string   `mapstructure:"stage" validate:"nonzero"`
	ID        string   `mapstructure:"id,omitempty" validate:"nonzero,nowhitespace"`
	Script    []string `mapstructure:"script" validate:"nonzero"`
	If        string   `mapstructure:"if,omitempty"`
	Clean     bool     `mapstructure:"clean,omitempty"`
	Env       []Env    `mapstructure:"env,omitempty"`
	Artifacts []string `mapstructure:"artifacts,omitempty"`
	Image     string   `mapstructure:"image,omitempty"`
	Needs     string   `mapstructure:"needs,omitempty" validate:"nowhitespace"`
}

type Env struct {
	Name  string `mapstructure:"name" validate:"nonzero,nowhitespace"`
	Value string `mapstructure:"value" validate:"nonzero"`
}

type Workflow struct {
	ID          string  `mapstructure:"id" validate:"nonzero,nowhitespace"`
	Image       string  `mapstructure:"image" validate:"nonzero,nowhitespace"`
	Description string  `mapstructure:"description"`
	Env         []Env   `mapstructure:"env"`
	Input       []Input `mapstructure:"input"`
	Stages      []Stage `mapstructure:"stages" validate:"nonzero"`
}

func GenEnv(e []Env) []string {
	envs := make([]string, len(e))
	for i, v := range e {
		envs[i] = fmt.Sprintf("%s=%s", v.Name, v.Value)
	}
	return envs
}
func GenEnvFromArgs(e []Input) []Env {
	envs := make([]Env, len(e))
	for i, v := range e {
		envs[i].Name = v.Name
		envs[i].Value = v.Default
	}
	return envs
}

func ImageExists(image string, ctx context.Context, cli *client.Client) bool {

	images, err := cli.ImageList(ctx, types.ImageListOptions{})
	logger.HandleErr(err)

	for _, extImage := range images {
		if strings.Join(extImage.RepoTags[:], ":") == image {
			// logger.Info("Image Already Exists -", image)
			return true
		}
	}
	return false
}

func ContainerClean(id string, ctx context.Context, cli *client.Client) {

	err := cli.ContainerRemove(ctx, id, types.ContainerRemoveOptions{})

	logger.HandleErr(err)
}

func FindVolume(name string, ctx context.Context, cli *client.Client) (volume *types.Volume, err error) {
	volumes, err := cli.VolumeList(ctx, filters.NewArgs())

	if err != nil {
		return nil, err
	}

	for _, v := range volumes.Volumes {
		if v.Name == name {
			return v, nil
		}
	}
	return nil, nil
}
func RemoveVolume(name string, ctx context.Context, cli *client.Client) (removed bool, err error) {
	vol, err := FindVolume(name, ctx, cli)

	if err != nil {
		return false, err
	}

	if vol == nil {
		return false, nil
	}

	err = cli.VolumeRemove(ctx, name, true)

	if err != nil {
		return false, err
	}

	return true, nil
}

func RunStage(s Stage, ctx context.Context, cli *client.Client, envs []Env, globalImage string, volume types.Volume, dir string, volumeOutput types.Volume, dirOutput string, LwWhite *logger.MyLogWriter) {

	PullImage(s.Image, ctx, cli)
	if s.Image != "" {
		globalImage = s.Image
	}

	hostConfig := container.HostConfig{}

	//	hostConfig.Mounts = make([]mount.Mount,0);

	var mounts []mount.Mount

	// for _, volume := range volumes {
	mounts = append(mounts, mount.Mount{
		Type:   mount.TypeVolume,
		Source: volume.Name,
		Target: "/app",
	}, mount.Mount{
		Type:   mount.TypeVolume,
		Source: volumeOutput.Name,
		Target: "/output",
	})
	// }

	hostConfig.Mounts = mounts
	allEnvs := GenEnv(envs)
	LwWhite.Write([]byte(fmt.Sprintf("Running Stage with the following variables: %s\n", allEnvs)))
	allEnvs = append(allEnvs, []string{fmt.Sprintf("OUTPUT=/output/output")}...)
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:      globalImage,
		Env:        allEnvs,
		Cmd:        s.Script,
		WorkingDir: "/app",
		Tty:        false,
	}, &hostConfig, nil, nil, "")
	logger.HandleErr(err)

	defer extractArtifacts(dir, s)
	defer ContainerClean(resp.ID, ctx, cli)

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		logger.HandleErr(err)
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	logger.HandleErr(err)

	stdcopy.StdCopy(LwWhite, LwWhite, out)

}

func PullImage(image string, ctx context.Context, cli *client.Client) {
	if image != "" {
		if !ImageExists(image, ctx, cli) {
			logger.Info("Pulling Image", image)
			reader, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
			logger.HandleErr(err)
			defer reader.Close()
			logger.Info("Image", image, "Pulled Successfully")
		}
		// io.Copy(os.Stdout, reader)
	}
}

func ReadPropertiesFile(filename string) ([]Env, error) {
	config := []Env{}

	if len(filename) == 0 {
		return config, nil
	}
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if equal := strings.Index(line, "="); equal >= 0 {
			if key := strings.TrimSpace(line[:equal]); len(key) > 0 {
				value := ""
				if len(line) > equal {
					value = strings.TrimSpace(line[equal+1:])
				}
				config = append(config, Env{Name: key, Value: value})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
		return nil, err
	}

	return config, nil
}

func Engine(cli *client.Client, ctx context.Context, w Workflow, sID string, vol types.Volume, dir string, allOutputs map[string][]Env, wg *sync.WaitGroup, skippedStages *[]string) {
	defer wg.Done()
	idx := slices.IndexFunc(w.Stages, func(c Stage) bool { return c.ID == sID })
	stage := w.Stages[idx]
	volOutput, dirOutput := createVolume(cli, ctx)
	outputPath := path.Join(dirOutput, "output")
	_, err := os.Create(outputPath)
	logger.HandleErr(err)

	defer RemoveVolume(volOutput.Name, ctx, cli)
	defer os.RemoveAll(dirOutput)

	allEnvs := append(w.Env, stage.Env...)
	allEnvs = append(allEnvs, GenEnvFromArgs(w.Input)...)
	needSplit := strings.Split(stage.Needs, ",")
	if stage.Needs != "" {

		for _, v := range needSplit {
			if val, ok := allOutputs[v]; ok {
				allEnvs = append(allEnvs, val...)
			}
		}
	}
	LwWhite := logger.NewLogWriter(func(str string, color color.Attribute) {
		logger.Custom(color, fmt.Sprintf("[%s:%s] %s", stage.Stage, stage.ID, str))
	}, color.FgWhite)

	LwCrossed := log.New(logger.NewLogWriter(func(str string, col color.Attribute) {
		colFuc := color.New(col).SprintFunc()
		white := color.New(color.CrossedOut).SprintFunc()
		logger.Free(white(fmt.Sprintf("[%s:%s] ", stage.Stage, stage.ID), colFuc(str)))
	}, color.BgYellow), "", 0)

	if !evaluateCondition(stage.If, allEnvs, LwWhite) {
		skippedStages = append(skippedStages, stage.ID)
		LwCrossed.Println("Stage Skipped due to IF condition")
	} else {
		toSkip := false
		for _, skipped := range skippedStages {
			for _, need := range needSplit {
				if need == skipped {
					toSkip = true
				}
			}
		}
		if toSkip {
			skippedStages = append(skippedStages, stage.ID)
			LwCrossed.Println("Stage Skipped due to needed stage skipped")
		} else {
			if stage.Clean {
				volClean, dirClean := createVolume(cli, ctx)
				RunStage(stage, ctx, cli, allEnvs, w.Image, volClean, dirClean, volOutput, dirOutput, LwWhite)
			} else {
				RunStage(stage, ctx, cli, allEnvs, w.Image, vol, dir, volOutput, dirOutput, LwWhite)
			}
		}

	}

	outputMap, err := ReadPropertiesFile(outputPath)
	logger.HandleErr(err)
	allOutputs[stage.ID] = outputMap
}

func createVolume(cli *client.Client, ctx context.Context) (vol types.Volume, dir string) {
	dir, err := os.MkdirTemp("", "temp")
	logger.HandleErr(err)

	vol, err2 := cli.VolumeCreate(ctx, volume.VolumeCreateBody{
		Driver: "local",
		DriverOpts: map[string]string{
			"type":   "none",
			"device": dir,
			"o":      "bind",
		},
	})
	logger.HandleErr(err2)
	return vol, dir
}

func trimFirstRune(s string) string {
	_, i := utf8.DecodeRuneInString(s)
	return s[i:]
}

func getVariablesFromExpression(condition string) []string {
	fields := strings.Fields(condition)
	varList := make([]string, 0)
	for _, v := range fields {
		if strings.HasPrefix(v, "$") {
			varList = append(varList, strings.ReplaceAll(trimFirstRune(v), " ", ""))
		}
	}
	return varList
}

func evaluateCondition(condition string, availableValues []Env, LwWhite *logger.MyLogWriter) bool {
	if condition != "" {
		LwWhite.Write([]byte(fmt.Sprintf("Evaluating If Statement: %s, with the following variables: %s\n", condition, availableValues)))

		varList := getVariablesFromExpression(condition)

		parameters := make(map[string]interface{}, len(varList))

		for _, v := range varList {
			idx := slices.IndexFunc(availableValues, func(c Env) bool { return c.Name == v })
			if idx == -1 {
				// Not all variables can be populated. Thus the If statement is void.
				return false
			} else {
				parameters[v] = availableValues[idx].Value
			}
		}

		expression, err := govaluate.NewEvaluableExpression(strings.ReplaceAll(condition, "$", ""))
		logger.HandleErr(err)

		result, err := expression.Evaluate(parameters)
		logger.HandleErr(err)
		return result.(bool)
	}
	return true
}

func extractArtifacts(path string, s Stage) {
	white := color.New(color.FgWhite).SprintFunc()

	LwOperation := log.New(logger.NewLogWriter(func(str string, col color.Attribute) {
		colFuc := color.New(col).SprintFunc()
		logger.Free(white(fmt.Sprintf("[%s:%s] ", s.Stage, s.ID), colFuc(str)))
	}, color.FgYellow), "", 0)
	LwSuccess := log.New(logger.NewLogWriter(func(str string, col color.Attribute) {
		colFuc := color.New(col).SprintFunc()
		logger.Free(white(fmt.Sprintf("[%s:%s] ", s.Stage, s.ID), colFuc(str)))
	}, color.FgGreen), "", 0)
	LwError := log.New(logger.NewLogWriter(func(str string, col color.Attribute) {
		colFuc := color.New(col).SprintFunc()
		logger.Free(white(fmt.Sprintf("[%s:%s] ", s.Stage, s.ID), colFuc(str)))
	}, color.FgRed), "", 0)

	for _, v := range s.Artifacts {
		fullPath := filepath.Join(path, v)
		fi, err := os.Stat(fullPath)
		if err != nil {
			LwError.Println(err.Error())
			return
		}
		current, _ := os.Getwd()
		to := filepath.Join(current, v)
		LwOperation.Println("Copying", v, "To", to)

		switch mode := fi.Mode(); {
		case mode.IsDir():
			// do directory stuff
			err = cp.Copy(fullPath, to)
			if err != nil {
				LwError.Println(err.Error())
			} else {
				LwSuccess.Println("Copied", v, "To", to)
			}
		case mode.IsRegular():
			// do file stuff
			// Read all content of src to data
			data, _ := ioutil.ReadFile(fullPath)
			// Write data to dst
			err = ioutil.WriteFile(to, data, 0644)
			if err != nil {
				LwError.Println(err.Error())
			} else {
				LwSuccess.Println("Copied", v, "To", to)
			}

		}
	}
}
