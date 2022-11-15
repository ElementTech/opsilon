package list

import (
	"github.com/jatalocks/opsilon/internal/config"
	"github.com/jatalocks/opsilon/internal/get"
	"github.com/jatalocks/opsilon/internal/logger"
)

func List(repoList []string) {
	w, err := get.GetWorkflowsForRepo(repoList)
	logger.HandleErr(err)
	config.PrintWorkflows(w)
}
