package config

import (
	"github.com/jatalocks/opsilon/internal/logger"
	"github.com/spf13/viper"
)

type Location struct {
	Path string `mapstructure:"path" validate:"nonzero"`
	Type string `mapstructure:"type" validate:"nonzero"`
}

type Repo struct {
	Name        string   `mapstructure:"name" validate:"nonzero"`
	Description string   `mapstructure:"description"`
	Location    Location `mapstructure:"location" validate:"nonzero"`
}

type RepoFile struct {
	Repositories []Repo `mapstructure:"repositories" validate:"nonzero"`
}

var C RepoFile

func GetConfig() []Repo {
	err2 := viper.Unmarshal(&C)
	logger.HandleErr(err2)
	return C.Repositories
}

func GetRepoList() []string {
	temp := []string{}
	err2 := viper.Unmarshal(&C)
	logger.HandleErr(err2)
	for _, r := range C.Repositories {
		temp = append(temp, r.Name)
	}
	return temp
}
