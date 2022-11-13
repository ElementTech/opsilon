package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/jatalocks/opsilon/internal/get"
	"github.com/jatalocks/opsilon/internal/logger"
)

func ConfigExists() bool {
	dirname, err := os.UserHomeDir()
	logger.HandleErr(err)
	if _, err := os.Stat(dirname + "/" + ".opsilon" + "/opsilon.yaml"); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func GetConfig(f string) []get.Action {
	if f == "" {
		if ConfigExists() {
			dirname, err := os.UserHomeDir()
			logger.HandleErr(err)
			return get.List(dirname + "/" + ".opsilon" + "/opsilon.yaml")
		} else {
			cyan := color.New(color.FgCyan).SprintFunc()
			bold := color.New(color.Bold).SprintFunc()
			fmt.Printf("%s Please run %s or pass a file using %s", bold("No opsilon configuration exists."), cyan("opsilon configure"), cyan("-f/--file"))
		}
	} else {
		return get.List(f)
	}
	return []get.Action{}
}
