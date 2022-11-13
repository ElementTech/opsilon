package get

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/jatalocks/opsilon/internal/engine"
	"github.com/jatalocks/opsilon/internal/logger"
	"github.com/manifoldco/promptui"
	"gopkg.in/yaml.v2"
)

type Location struct {
	Path string `yaml:"path"`
	Type string `yaml:"type"`
}

type Action struct {
	Name     string          `yaml:"name"`
	Location Location        `yaml:"location"`
	Workflow engine.Workflow `yaml:"workflow,omitempty"`
}

func getWorkflow(location Location) *engine.Workflow {
	if location.Type == "file" {
		yfile, err := ioutil.ReadFile(location.Path)

		logger.HandleErr(err)
		data := engine.Workflow{}

		err2 := yaml.Unmarshal(yfile, &data)
		if err2 != nil {

			log.Fatal(err2)
		}
		return &data
	}
	return &engine.Workflow{}
}

func List(file string) []Action {
	yfile, err := ioutil.ReadFile(file)

	logger.HandleErr(err)

	data := []Action{}

	err2 := yaml.Unmarshal(yfile, &data)
	if err2 != nil {

		log.Fatal(err2)
	}
	for i, v := range data {
		data[i].Workflow = *getWorkflow(v.Location)
	}

	return data
}

func Confirm(act Action) (bool, error) {
	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("Run %v", act.Name),
		IsConfirm: true,
		Default:   "y",
	}
	validate := func(s string) error {
		if len(s) == 1 && strings.Contains("YyNn", s) || prompt.Default != "" && len(s) == 0 {
			return nil
		}
		return errors.New("invalid input")
	}
	prompt.Validate = validate

	_, err := prompt.Run()
	confirmed := !errors.Is(err, promptui.ErrAbort)
	if err != nil && confirmed {
		fmt.Println("ERROR: ", err)
		return false, err
	}

	return confirmed, nil
}
