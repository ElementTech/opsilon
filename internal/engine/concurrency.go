package engine

import (
	"fmt"
	"strings"

	"github.com/kendru/darwin/go/depgraph"
)

func workflowToGraph(g *depgraph.Graph, w Workflow) {
	for _, s := range w.Stages {
		g.DependOn(s.ID, s.Needs)
	}
}

//	func stageRunner(c chan []Stage) {
//		for {
//			msg := <-c
//			fmt.Println(msg)
//			time.Sleep(time.Second * 1)
//		}
//	}
func ToGraph(w Workflow) {
	// var c chan string = make(chan string)

	// go pinger(c)
	// go printer(c)

	g := depgraph.New()
	workflowToGraph(g, w)

	for i, layer := range g.TopoSortedLayers() {
		fmt.Printf("%d: %s\n", i, strings.Join(layer, ", "))
	}
	fmt.Println(g.TopoSortedLayers())
}

// func generateGraph(w Workflow) {
// 	var c chan []Stage = make(chan []Stage)
// 	paraArr := make([]Stage, 0)
// 	for _, s := range w.Stages {
// 		if s.Needs != "" {
// 			paraArr = append(paraArr, s)
// 		}
// 	}
// }
