package repo

import (
	"github.com/jatalocks/opsilon/internal/config"
)

func List() {
	config.PrintRepos(config.GetConfig())
}
