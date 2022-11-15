package utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/jatalocks/opsilon/internal/config"
	"github.com/jatalocks/opsilon/internal/engine"
	"github.com/jatalocks/opsilon/internal/logger"
	"github.com/jatalocks/opsilon/internal/validate"
	"github.com/manifoldco/promptui"
	"gopkg.in/yaml.v3"
)

func getWorkflows(location config.Location) *[]engine.Workflow {
	data := []engine.Workflow{}
	if location.Type == "folder" {
		if location.Path[0:1] == "/" {

			err := filepath.Walk(location.Path,
				func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}

					if !info.IsDir() {
						yfile, err := ioutil.ReadFile(path)
						logger.HandleErr(err)
						temp := engine.Workflow{}
						err2 := yaml.Unmarshal(yfile, &temp)
						logger.HandleErr(err2)
						data = append(data, temp)
					}

					return nil
				})
			if err != nil {
				logger.Fatal(err)
			}

		}
	}

	// else {

	// 	yfile, err2 := ioutil.ReadFile(path.Join(path.Dir(viper.ConfigFileUsed()), location.Path))
	// 	logger.HandleErr(err2)
	// 	err3 := yaml.Unmarshal(yfile, &data)
	// 	logger.HandleErr(err3)
	// }

	// } else if location.Type == "url" {
	// 	resp, err := http.Get(location.Path)
	// 	logger.HandleErr(err)
	// 	defer resp.Body.Close()
	// 	buf := new(bytes.Buffer)
	// 	buf.ReadFrom(resp.Body)
	// 	err3 := yaml.Unmarshal(buf.Bytes(), &data)
	// 	logger.HandleErr(err3)
	// }
	return &data
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func GetWorkflowsForRepo(repoList []string) []engine.Workflow {
	data := config.GetConfig()
	workflowArray := []engine.Workflow{}
	skipRepoCheck := false
	if len(repoList) == 0 {
		skipRepoCheck = true
	}
	for _, v := range data {
		if skipRepoCheck {
			logger.Info("Repository", v.Name)
			w := *getWorkflows(v.Location)
			validate.ValidateWorkflows(&w)
			workflowArray = append(workflowArray, w...)
		} else {
			if stringInSlice(v.Name, repoList) {
				logger.Info("Repository", v.Name)
				w := *getWorkflows(v.Location)
				validate.ValidateWorkflows(&w)
				workflowArray = append(workflowArray, w...)
			}
		}
	}

	return workflowArray
}

func Confirm(act engine.Workflow) (bool, error) {
	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("Run %v", act.ID),
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
