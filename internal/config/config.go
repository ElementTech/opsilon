package config

import (
	"errors"
	"log"
	"os"

	"github.com/jatalocks/opsilon/internal/get"
)

func ConfigExists() bool {
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	if _, err := os.Stat(dirname + "/" + ".opsilon" + "/opsilon.yaml"); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func GetConfig(f string) get.ActionFile {
	if f == "" {
		dirname, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		return get.List(dirname + "/" + ".opsilon" + "/opsilon.yaml")
	} else {
		return get.List(f)
	}
}
