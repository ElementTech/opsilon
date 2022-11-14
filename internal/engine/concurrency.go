package engine

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/jatalocks/opsilon/internal/logger"
	"github.com/kendru/darwin/go/depgraph"
)

func workflowToGraph(g *depgraph.Graph, w Workflow) {
	for _, s := range w.Stages {
		needSplit := strings.Split(s.Needs, ",")
		for _, v := range needSplit {
			g.DependOn(s.ID, v)
		}
	}
}

func runStageGroup(wg *sync.WaitGroup, stageIDs []string, cli *client.Client, ctx context.Context, w Workflow, vol types.Volume, dir string, allOutputs map[string][]Env) {
	for _, id := range stageIDs {
		go Engine(cli, ctx, w, id, vol, dir, allOutputs, wg)
	}
}
func ToGraph(w Workflow) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	logger.HandleErr(err)
	defer cli.Close()

	PullImage(w.Image, ctx, cli)

	vol, dir := createVolume(cli, ctx)

	defer RemoveVolume(vol.Name, ctx, cli)
	defer os.RemoveAll(dir)

	allOutputs := make(map[string][]Env, 0)

	wg := new(sync.WaitGroup)
	g := depgraph.New()
	workflowToGraph(g, w)
	for _, layer := range g.TopoSortedLayers() {
		if (len(layer) > 0) && (layer[0] != "") {
			fmt.Printf("Running in Parallel: %s\n", strings.Join(layer, ", "))
			wg.Add(len(layer))
			go runStageGroup(wg, layer, cli, ctx, w, vol, dir, allOutputs)
			wg.Wait()
		}
	}
}
