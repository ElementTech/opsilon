package engine

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
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
	"github.com/jatalocks/opsilon/internal/db"
	"github.com/jatalocks/opsilon/internal/internaltypes"
	"github.com/jatalocks/opsilon/internal/logger"
	"github.com/mitchellh/hashstructure/v2"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"
)

func GenEnv(e []internaltypes.Env) []string {
	envs := make([]string, len(e))
	for i, v := range e {
		envs[i] = fmt.Sprintf("%s=%s", v.Name, v.Value)
	}
	return envs
}

func GenEnvFromArgs(e []internaltypes.Input) []internaltypes.Env {
	envs := make([]internaltypes.Env, len(e))
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

// CopyDir copies the content of src to dst. src should be a full path.
func Copy(src, dst string) error {

	return filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// copy to this path
		outpath := filepath.Join(dst, strings.TrimPrefix(path, src))

		if info.IsDir() {
			os.MkdirAll(outpath, info.Mode())
			return nil // means recursive
		}

		// handle irregular files
		if !info.Mode().IsRegular() {
			switch info.Mode().Type() & os.ModeType {
			case os.ModeSymlink:
				link, err := os.Readlink(path)
				if err != nil {
					return err
				}
				return os.Symlink(link, outpath)
			}
			return nil
		}

		// copy contents of regular file efficiently

		// open input
		in, _ := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()

		// create output
		fh, err := os.Create(outpath)
		if err != nil {
			return err
		}
		defer fh.Close()

		// make it the same
		fh.Chmod(info.Mode())

		// copy content
		_, err = io.Copy(fh, in)
		return err
	})
}
func LoadImportsIntoStage(s internaltypes.Stage, targetDir, runid string, w internaltypes.Workflow) {
	current, _ := os.Getwd()
	for _, v := range s.Import {
		for _, a := range v.Artifacts {
			from := filepath.Join(current, "artifacts", w.Repo, w.ID, runid, v.From, a)
			to := filepath.Join(targetDir, a)
			fmt.Println("Copying", from, "To", to)
			err := Copy(from, to)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func RunStage(s internaltypes.Stage, ctx context.Context, cli *client.Client, envs []internaltypes.Env, globalImage string, volume types.Volume, dir string, volumeOutput types.Volume, dirOutput string, LwWhite *logger.MyLogWriter, LwRed *logger.MyLogWriter, runid string, w internaltypes.Workflow) bool {
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
	LoadImportsIntoStage(s, dir, runid, w)

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

	defer ExtractArtifacts(dir, s, runid, w)
	defer ContainerClean(resp.ID, ctx, cli)

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}
	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNextExit)

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	logger.HandleErr(err)
	stdcopy.StdCopy(LwWhite, LwRed, out)

	select {
	case err := <-errCh:
		logger.Error(err.Error())
	case status := <-statusCh:
		// stat, err := cli.ContainerInspect(ctx, resp.ID)
		// if err != nil {
		// 	fmt.Println(err)
		// }
		// fmt.Println("STAT", stat.State)
		return status.StatusCode == 0
	}

	return false
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

func ReadPropertiesFile(filename string) ([]internaltypes.Env, error) {
	config := []internaltypes.Env{}

	if len(filename) == 0 {
		return config, nil
	}
	file, err := os.Open(filename)
	if err != nil {
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
				config = append(config, internaltypes.Env{Name: key, Value: value})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return config, nil
}

func PrepareStage(wEnv []internaltypes.Env, sEnv []internaltypes.Env, inputs []internaltypes.Input, needs string, allOutputs map[string][]internaltypes.Env, stage string, id string, result *internaltypes.Result, runid, hash string) ([]internaltypes.Env, []string, *logger.MyLogWriter, *log.Logger, *logger.MyLogWriter) {
	allEnvs := append(wEnv, sEnv...)
	allEnvs = append(allEnvs, GenEnvFromArgs(inputs)...)
	needSplit := strings.Split(needs, ",")
	if needs != "" {
		for _, v := range needSplit {
			if val, ok := allOutputs[v]; ok {
				allEnvs = append(allEnvs, val...)
			}
		}
	}
	LwWhite := logger.NewLogWriter(func(str string, color color.Attribute) {
		logger.Custom(color, fmt.Sprintf("[%s:%s] %s", stage, id, str))
		if viper.GetBool("database") {
			for _, log := range result.Logs {
				go db.InsertOne("logs", internaltypes.RunLog{Log: log, Stage: id, RunID: runid, Workflow: hash, CreatedDate: time.Now(), UpdatedDate: time.Now()})
			}
		}
		result.Logs = append(result.Logs, fmt.Sprintf("[%s:%s] %s", stage, id, str))
	}, color.FgWhite)

	LwRed := logger.NewLogWriter(func(str string, color color.Attribute) {
		logger.Custom(color, fmt.Sprintf("[%s:%s] %s", stage, id, str))
		if viper.GetBool("database") {
			for _, log := range result.Logs {
				go db.InsertOne("logs", internaltypes.RunLog{Log: log, Stage: id, RunID: runid, Workflow: hash, CreatedDate: time.Now(), UpdatedDate: time.Now()})
			}
		}
		result.Logs = append(result.Logs, fmt.Sprintf("[%s:%s] %s", stage, id, str))
	}, color.FgRed)

	LwCrossed := log.New(logger.NewLogWriter(func(str string, col color.Attribute) {
		colFuc := color.New(col).SprintFunc()
		white := color.New(color.CrossedOut).SprintFunc()
		logger.Free(white(fmt.Sprintf("[%s:%s] ", stage, id), colFuc(str)))
		if viper.GetBool("database") {
			for _, log := range result.Logs {
				go db.InsertOne("logs", internaltypes.RunLog{Log: log, Stage: id, RunID: runid, Workflow: hash, CreatedDate: time.Now(), UpdatedDate: time.Now()})
			}
		}

		result.Logs = append(result.Logs, fmt.Sprintf("[%s:%s] %s", stage, id, str))
	}, color.BgYellow), "", 0)

	return allEnvs, needSplit, LwWhite, LwCrossed, LwRed
}

func Engine(cli *client.Client, ctx context.Context, w internaltypes.Workflow, sID string, allOutputs map[string][]internaltypes.Env, wg *sync.WaitGroup, skippedStages *[]string, results chan internaltypes.Result, runid string) {
	defer wg.Done()
	idx := slices.IndexFunc(w.Stages, func(c internaltypes.Stage) bool { return c.ID == sID })
	stage := w.Stages[idx]
	result := internaltypes.Result{Stage: stage}
	volOutput, dirOutput := CreateVolume(cli, ctx)
	outputPath := path.Join(dirOutput, "output")
	_, err := os.Create(outputPath)
	logger.HandleErr(err)

	defer RemoveVolume(volOutput.Name, ctx, cli)
	defer os.RemoveAll(dirOutput)

	tempW := w
	tempW.Input = []internaltypes.Input{}
	hash, err := hashstructure.Hash(tempW, hashstructure.FormatV2, nil)
	strHash := fmt.Sprint(hash)
	logger.HandleErr(err)

	allEnvs, needSplit, LwWhite, LwCrossed, LwRed := PrepareStage(w.Env, stage.Env, w.Input, stage.Needs, allOutputs, stage.Stage, stage.ID, &result, runid, strHash)
	if !EvaluateCondition(stage.If, allEnvs, LwWhite) {
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
			vol, dir := CreateVolume(cli, ctx)
			success := RunStage(stage, ctx, cli, allEnvs, w.Image, vol, dir, volOutput, dirOutput, LwWhite, LwRed, runid, w)
			result.Result = success
		}

	}

	outputMap, err := ReadPropertiesFile(outputPath)
	logger.HandleErr(err)
	allOutputs[stage.ID] = outputMap
	result.Outputs = outputMap
	results <- result
}

func CreateVolume(cli *client.Client, ctx context.Context) (vol types.Volume, dir string) {
	dir, err := os.MkdirTemp("", "temp")
	logger.HandleErr(err)

	// if mount {
	// 	wd, err := os.Getwd()
	// 	logger.HandleErr(err)
	// 	err2 := cp.Copy(wd, dir)
	// 	fmt.Println(err2) // nil
	// }

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

func EvaluateCondition(condition string, availableValues []internaltypes.Env, LwWhite *logger.MyLogWriter) bool {
	if condition != "" {
		LwWhite.Write([]byte(fmt.Sprintf("Evaluating If Statement: %s, with the following variables: %s\n", condition, availableValues)))

		varList := getVariablesFromExpression(condition)

		parameters := make(map[string]interface{}, len(varList))

		for _, v := range varList {
			idx := slices.IndexFunc(availableValues, func(c internaltypes.Env) bool { return c.Name == v })
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

func ExtractArtifacts(path string, s internaltypes.Stage, runid string, w internaltypes.Workflow) {
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
		current, _ := os.Getwd()
		to := filepath.Join(current, "artifacts", w.Repo, w.ID, runid, s.ID, v)
		os.MkdirAll(filepath.Join(current, "artifacts", w.Repo, w.ID, runid, s.ID), 0755)
		LwOperation.Println("Copying", v, "To", to)
		err := Copy(fullPath, to)
		if err != nil {
			LwError.Println(err.Error())
		} else {
			LwSuccess.Println("Copied", v, "To", to)
		}
	}
}
