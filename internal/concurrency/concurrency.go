package concurrency

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/jatalocks/opsilon/internal/config"
	"github.com/jatalocks/opsilon/internal/engine"
	"github.com/jatalocks/opsilon/internal/internaltypes"
	"github.com/jatalocks/opsilon/internal/kubengine"
	"github.com/jatalocks/opsilon/internal/logger"
	"github.com/kendru/darwin/go/depgraph"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	v1 "k8s.io/api/core/v1"
)

func workflowToGraph(g *depgraph.Graph, w internaltypes.Workflow) {
	for _, s := range w.Stages {
		needSplit := strings.Split(s.Needs, ",")
		for _, v := range needSplit {
			g.DependOn(s.ID, v)
		}
	}
}

func runStageGroupDocker(wg *sync.WaitGroup, stageIDs []string, cli *client.Client, ctx context.Context, w internaltypes.Workflow, vol types.Volume, dir string, allOutputs map[string][]internaltypes.Env, skippedStages *[]string, results chan internaltypes.Result) {
	for _, id := range stageIDs {
		go engine.Engine(cli, ctx, w, id, vol, dir, allOutputs, wg, skippedStages, results)
	}
}

// cli, wg, layer, cli, ctx, w, vol, allOutputs, &skippedStages
func runStageGroupKubernetes(cli *kubengine.Client, wg *sync.WaitGroup, stageIDs []string, ctx context.Context, w internaltypes.Workflow, vol string, claim *v1.PersistentVolumeClaim, allOutputs map[string][]internaltypes.Env, skippedStages *[]string, results chan internaltypes.Result) {
	for _, id := range stageIDs {
		go cli.KubeEngine(wg, id, ctx, w, vol, claim, allOutputs, skippedStages, results)
	}
}

func ToGraph(w internaltypes.Workflow, c echo.Context) {
	skippedStages := make([]string, 0)
	ctx := context.Background()
	k8s := viper.GetBool("kubernetes")
	fmt.Println("K8S", k8s)
	results := make(chan internaltypes.Result)
	resultsArray := []internaltypes.Result{}
	if !k8s {
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		logger.HandleErr(err)
		defer cli.Close()
		engine.PullImage(w.Image, ctx, cli)

		vol, dir := engine.CreateVolume(cli, ctx, w.Mount)

		defer engine.RemoveVolume(vol.Name, ctx, cli)
		defer os.RemoveAll(dir)

		allOutputs := make(map[string][]internaltypes.Env, 0)

		wg := new(sync.WaitGroup)
		g := depgraph.New()
		workflowToGraph(g, w)
		for _, layer := range g.TopoSortedLayers() {
			if (len(layer) > 0) && (layer[0] != "") {
				fmt.Printf("Running in Parallel: %s\n", strings.Join(layer, ", "))
				wg.Add(len(layer))
				go func() {
					for str := range results {
						if str.Result {
							logger.Success("Stage", str.Stage.ID, "Success")
						} else {
							if str.Skipped {
								logger.Operation("Stage", str.Stage.ID, "Skipped")
							} else {
								logger.Error("Stage", str.Stage.ID, "Failed")
							}
						}
						if c != nil {
							streamResultToEchoContext(c, str)
						}
						resultsArray = append(resultsArray, str)
					}
				}()
				go runStageGroupDocker(wg, layer, cli, ctx, w, vol, dir, allOutputs, &skippedStages, results)
				wg.Wait()
			}
		}
	} else {
		cli, err := kubengine.NewClient()
		logger.HandleErr(err)
		// defer cli.DeleteNamespace(ctx)
		vol, claim := cli.CreateVolume(ctx, w.Mount)
		allOutputs := make(map[string][]internaltypes.Env, 0)

		wg := new(sync.WaitGroup)
		g := depgraph.New()
		workflowToGraph(g, w)
		for _, layer := range g.TopoSortedLayers() {
			if (len(layer) > 0) && (layer[0] != "") {
				fmt.Printf("Running in Parallel: %s\n", strings.Join(layer, ", "))
				wg.Add(len(layer))
				go func() {
					for str := range results {
						if str.Result {
							logger.Success("Stage", str.Stage.ID, "Success")
						} else {
							if str.Skipped {
								logger.Operation("Stage", str.Stage.ID, "Skipped")
							} else {
								logger.Error("Stage", str.Stage.ID, "Failed")
							}
						}
						if c != nil {
							streamResultToEchoContext(c, str)
						}
						resultsArray = append(resultsArray, str)
					}
				}()
				go runStageGroupKubernetes(cli, wg, layer, ctx, w, vol, claim, allOutputs, &skippedStages, results)
				wg.Wait()
			}
		}
		cli.RemoveVolume(ctx, vol, claim)

	}
	config.PrintStageResults(resultsArray)
}

func streamResultToEchoContext(c echo.Context, result internaltypes.Result) error {
	enc := json.NewEncoder(c.Response())
	if err := enc.Encode(result); err != nil {
		return err
	}
	c.Response().Flush()
	// time.Sleep(1 * time.Second)
	return nil
}
