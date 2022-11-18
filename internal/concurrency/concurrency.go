package concurrency

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/jatalocks/opsilon/internal/engine"
	"github.com/jatalocks/opsilon/internal/kubengine"
	"github.com/jatalocks/opsilon/internal/logger"
	"github.com/kendru/darwin/go/depgraph"
	"github.com/spf13/viper"
	v1 "k8s.io/api/core/v1"
)

func workflowToGraph(g *depgraph.Graph, w engine.Workflow) {
	for _, s := range w.Stages {
		needSplit := strings.Split(s.Needs, ",")
		for _, v := range needSplit {
			g.DependOn(s.ID, v)
		}
	}
}

func runStageGroupDocker(wg *sync.WaitGroup, stageIDs []string, cli *client.Client, ctx context.Context, w engine.Workflow, vol types.Volume, dir string, allOutputs map[string][]engine.Env, skippedStages *[]string) {
	for _, id := range stageIDs {
		go engine.Engine(cli, ctx, w, id, vol, dir, allOutputs, wg, skippedStages)
	}
}

// cli, wg, layer, cli, ctx, w, vol, allOutputs, &skippedStages
func runStageGroupKubernetes(cli *kubengine.Client, wg *sync.WaitGroup, stageIDs []string, ctx context.Context, w engine.Workflow, vol *v1.PersistentVolume, allOutputs map[string][]engine.Env, skippedStages *[]string) {
	for _, id := range stageIDs {
		go cli.KubeEngine(wg, id, ctx, w, vol, allOutputs, skippedStages)
	}
}

func ToGraph(w engine.Workflow) {
	skippedStages := make([]string, 0)
	ctx := context.Background()
	k8s := viper.GetBool("kubernetes")

	if !k8s {
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		logger.HandleErr(err)
		defer cli.Close()
		engine.PullImage(w.Image, ctx, cli)

		vol, dir := engine.CreateVolume(cli, ctx, w.Mount)

		defer engine.RemoveVolume(vol.Name, ctx, cli)
		defer os.RemoveAll(dir)

		allOutputs := make(map[string][]engine.Env, 0)

		wg := new(sync.WaitGroup)
		g := depgraph.New()
		workflowToGraph(g, w)
		for _, layer := range g.TopoSortedLayers() {
			if (len(layer) > 0) && (layer[0] != "") {
				fmt.Printf("Running in Parallel: %s\n", strings.Join(layer, ", "))
				wg.Add(len(layer))
				go runStageGroupDocker(wg, layer, cli, ctx, w, vol, dir, allOutputs, &skippedStages)
				wg.Wait()
			}
		}
	} else {
		cli, err := kubengine.NewClient()
		logger.HandleErr(err)

		vol := cli.CreateVolume(ctx, w.Mount)
		defer cli.RemoveVolume(ctx, vol)

		allOutputs := make(map[string][]engine.Env, 0)

		wg := new(sync.WaitGroup)
		g := depgraph.New()
		workflowToGraph(g, w)
		for _, layer := range g.TopoSortedLayers() {
			if (len(layer) > 0) && (layer[0] != "") {
				fmt.Printf("Running in Parallel: %s\n", strings.Join(layer, ", "))
				wg.Add(len(layer))
				go runStageGroupKubernetes(cli, wg, layer, ctx, w, vol, allOutputs, &skippedStages)
				wg.Wait()
			}
		}
	}
}
