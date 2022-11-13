package engine

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

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
)

type Argument struct {
	Name     string `mapstructure:"name"`
	Value    string `mapstructure:"default,omitempty"`
	Optional bool   `mapstructure:"optional,omitempty"`
}

type Stage struct {
	Stage  string   `mapstructure:"stage"`
	Script []string `mapstructure:"script"`
	Rules  []struct {
		If    string `mapstructure:"if"`
		Verb  string `mapstructure:"verb"`
		Value string `mapstructure:"value"`
	} `mapstructure:"rules,omitempty"`
	Artifacts []string `mapstructure:"artifacts,omitempty"`
	Image     string   `mapstructure:"image,omitempty"`
}

type Env struct {
	Name  string `mapstructure:"name"`
	Value string `mapstructure:"value"`
}

type Workflow struct {
	ID          string     `mapstructure:"id"`
	Image       string     `mapstructure:"image"`
	Description string     `mapstructure:"description"`
	Env         []Env      `mapstructure:"env"`
	Input       []Argument `mapstructure:"input"`
	Stages      []Stage    `mapstructure:"stages"`
}

func GenEnv(e []Env) []string {
	envs := make([]string, len(e))
	for i, v := range e {
		envs[i] = fmt.Sprintf("%s=%s", v.Name, v.Value)
	}
	return envs
}
func GenEnvFromArgs(e []Argument) []string {
	envs := make([]string, len(e))
	for i, v := range e {
		envs[i] = fmt.Sprintf("%s=%s", v.Name, v.Value)
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

func RunStage(s Stage, ctx context.Context, cli *client.Client, envs []Env, inputs []Argument, globalImage string, volume types.Volume, dir string) {
	LwWhite := logger.NewLogWriter(func(str string, color color.Attribute) {
		logger.Custom(color, fmt.Sprintf("[%s] %s", s.Stage, str))
	}, color.FgWhite)
	PullImage(s.Image, ctx, cli)
	if s.Image != "" {
		globalImage = s.Image
	}

	hostConfig := container.HostConfig{}

	//	hostConfig.Mounts = make([]mount.Mount,0);

	var mounts []mount.Mount

	// for _, volume := range volumes {
	mount := mount.Mount{
		Type:   mount.TypeVolume,
		Source: volume.Name,
		Target: "/app",
	}
	mounts = append(mounts, mount)
	// }

	hostConfig.Mounts = mounts

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:      globalImage,
		Env:        append(GenEnv(envs), GenEnvFromArgs(inputs)...),
		Cmd:        s.Script,
		WorkingDir: "/app",
		Tty:        false,
	}, &hostConfig, nil, nil, "")
	logger.HandleErr(err)

	defer ContainerClean(resp.ID, ctx, cli)

	defer extractArtifacts(dir, s)

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

func Engine(w Workflow) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	logger.HandleErr(err)
	defer cli.Close()

	PullImage(w.Image, ctx, cli)

	dir, err := os.MkdirTemp("", "app")
	logger.HandleErr(err)

	vol, err2 := cli.VolumeCreate(ctx, volume.VolumeCreateBody{
		Driver: "local",
		DriverOpts: map[string]string{
			"type":   "none",
			"device": dir,
			"o":      "bind",
		},
	})
	if err2 != nil {
		panic(err2)
	}
	defer RemoveVolume(vol.Name, ctx, cli)
	defer os.RemoveAll(dir)
	for _, stage := range w.Stages {
		RunStage(stage, ctx, cli, w.Env, w.Input, w.Image, vol, dir)
	}

}

func extractArtifacts(path string, s Stage) {
	white := color.New(color.FgWhite).SprintFunc()

	LwOperation := log.New(logger.NewLogWriter(func(str string, col color.Attribute) {
		colFuc := color.New(col).SprintFunc()
		logger.Free(white(fmt.Sprintf("[%s] ", s.Stage), colFuc(str)))
	}, color.FgYellow), "", 0)
	LwSuccess := log.New(logger.NewLogWriter(func(str string, col color.Attribute) {
		colFuc := color.New(col).SprintFunc()
		logger.Free(white(fmt.Sprintf("[%s] ", s.Stage), colFuc(str)))
	}, color.FgGreen), "", 0)
	LwError := log.New(logger.NewLogWriter(func(str string, col color.Attribute) {
		colFuc := color.New(col).SprintFunc()
		logger.Free(white(fmt.Sprintf("[%s] ", s.Stage), colFuc(str)))
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
