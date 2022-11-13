package utils

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"

	"github.com/jatalocks/opsilon/internal/config"
	"github.com/jatalocks/opsilon/internal/engine"
	"github.com/jatalocks/opsilon/internal/logger"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

func getWorkflow(location config.Location) *engine.Workflow {
	data := engine.Workflow{}
	if location.Type == "file" {
		if location.Path[0:1] == "/" {
			yfile, err := ioutil.ReadFile(location.Path)
			logger.HandleErr(err)
			err2 := yaml.Unmarshal(yfile, &data)
			logger.HandleErr(err2)
		} else {

			yfile, err2 := ioutil.ReadFile(path.Join(path.Dir(viper.ConfigFileUsed()), location.Path))
			logger.HandleErr(err2)
			err3 := yaml.Unmarshal(yfile, &data)
			logger.HandleErr(err3)
		}
	} else if location.Type == "url" {
		resp, err := http.Get(location.Path)
		logger.HandleErr(err)
		defer resp.Body.Close()
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		err3 := yaml.Unmarshal(buf.Bytes(), &data)
		logger.HandleErr(err3)
	}
	return &data
}

func ConfigPopulateWorkflows() []config.Action {
	data := config.GetConfig()

	for i, v := range data {
		data[i].Workflow = *getWorkflow(v.Location)
	}
	return data
}

func Confirm(act config.Action) (bool, error) {
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
