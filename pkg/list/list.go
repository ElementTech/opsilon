package list

import (
	"github.com/jatalocks/opsilon/internal/config"
	"github.com/jatalocks/opsilon/internal/utils"
)

func List(repoList []string) {
	config.PrintWorkflows(utils.GetWorkflowsForRepo(repoList))
}
