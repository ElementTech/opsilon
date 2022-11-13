package config

import (
	"github.com/jatalocks/opsilon/internal/engine"
	"github.com/jatalocks/opsilon/internal/logger"
	"github.com/spf13/viper"
)

type Location struct {
	Path string `mapstructure:"path"`
	Type string `mapstructure:"type"`
}

type Action struct {
	Name     string          `mapstructure:"name"`
	Location Location        `mapstructure:"location"`
	Workflow engine.Workflow `mapstructure:"workflow,omitempty"`
}

type ActionFile struct {
	Actions []Action `mapstructure:"workflows"`
}

var C ActionFile

func GetConfig() []Action {
	err2 := viper.Unmarshal(&C)
	logger.HandleErr(err2)
	return C.Actions
}
