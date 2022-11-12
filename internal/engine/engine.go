package engine

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
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

func RunStage(s Stage, ctx context.Context, cli *client.Client, envs []Env, inputs []Argument, globalImage string) {
	if s.Image != "" {
		reader, err := cli.ImagePull(ctx, s.Image, types.ImagePullOptions{})
		if err != nil {
			panic(err)
		}

		defer reader.Close()

		globalImage = s.Image
	}
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: globalImage,
		Env:   append(GenEnv(envs), GenEnvFromArgs(inputs)...),
		Cmd:   s.Script,
		Tty:   false,
	}, nil, nil, nil, "")
	if err != nil {
		panic(err)
	}

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
}

func Engine(w Workflow) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	if w.Image != "" {
		reader, err := cli.ImagePull(ctx, w.Image, types.ImagePullOptions{})
		if err != nil {
			panic(err)
		}
		defer reader.Close()
		// io.Copy(os.Stdout, reader)
	}

	for _, stage := range w.Stages {
		RunStage(stage, ctx, cli, w.Env, w.Input, w.Image)
	}

}
