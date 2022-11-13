package configure

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/jatalocks/opsilon/internal/get"
	"github.com/jatalocks/opsilon/internal/logger"
	"gopkg.in/yaml.v2"
)

func Configure(file string) {
	actions := get.List(file)

	dirname, err := os.UserHomeDir()
	logger.HandleErr(err)

	data, err := yaml.Marshal(&actions)

	logger.HandleErr(err)

	path := dirname + "/" + ".opsilon"
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(path, os.ModePerm)
		logger.HandleErr(err)
	}

	err2 := ioutil.WriteFile(path+"/opsilon.yaml", data, 0644)

	if err2 != nil {

		log.Fatal(err2)
	}

	fmt.Println("Opsilon Configuration Written to", path+"/opsilon.yaml")
}
