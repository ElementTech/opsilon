package engine

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/jatalocks/opsilon/internal/log"
)

type Argument struct {
	Name     string `yaml:"name"`
	Value    string `yaml:"default"`
	Optional bool   `yaml:"optional"`
}

type Stage struct {
	Stage  string   `yaml:"stage"`
	Script []string `yaml:"script"`
	Rules  []struct {
		If    string `yaml:"if"`
		Verb  string `yaml:"verb"`
		Value string `yaml:"value"`
	} `yaml:"rules,omitempty"`
	Artifacts struct {
		Paths []string `yaml:"paths"`
	} `yaml:"artifacts,omitempty"`
	Image string `yaml:"image,omitempty"`
}

type Env struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type Workflow struct {
	ID          string     `yaml:"id"`
	Image       string     `yaml:"image"`
	Description string     `yaml:"description"`
	Env         []Env      `yaml:"env"`
	Input       []Argument `yaml:"input"`
	Stages      []Stage    `yaml:"stages"`
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
	if err != nil {
		panic(err)
	}

	for _, extImage := range images {
		if strings.Join(extImage.RepoTags[:], ":") == image {
			// log.Info("Image Already Exists -", image)
			return true
		}
	}
	return false
}

func ContainerClean(id string, ctx context.Context, cli *client.Client) {

	err := cli.ContainerRemove(ctx, id, types.ContainerRemoveOptions{})

	if err != nil {
		fmt.Printf("Unable to remove container %q: %q\n", id, err)
	}
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

func RunStage(s Stage, ctx context.Context, cli *client.Client, envs []Env, inputs []Argument, globalImage string, volume types.Volume) {
	PullImage(s.Image, ctx, cli)
	if s.Image != "" {
		globalImage = s.Image
	}
	log.Info("----", "Starting Stage", s.Stage, "----")

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
	if err != nil {
		panic(err)
	}

	defer ContainerClean(resp.ID, ctx, cli)

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	log.Info("----", "Stage Ended", s.Stage, "----")
}

func PullImage(image string, ctx context.Context, cli *client.Client) {
	if image != "" {
		if !ImageExists(image, ctx, cli) {
			log.Info("Pulling Image", image)
			reader, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
			if err != nil {
				panic(err)
			}
			defer reader.Close()
			log.Info("Image", image, "Pulled Successfully")
		}
		// io.Copy(os.Stdout, reader)
	}
}

func Engine(w Workflow) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	PullImage(w.Image, ctx, cli)
	vol, err2 := cli.VolumeCreate(ctx, volume.VolumeCreateBody{Driver: "local"})
	if err2 != nil {
		panic(err2)
	}
	defer RemoveVolume(vol.Name, ctx, cli)
	for _, stage := range w.Stages {
		RunStage(stage, ctx, cli, w.Env, w.Input, w.Image, vol)
	}

}
