package get

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/manifoldco/promptui"
	"gopkg.in/yaml.v2"
)

type ActionFile struct {
	Actions []Action `yaml:"actions"`
}

type Action struct {
	Name string     `yaml:"name"`
	Help string     `yaml:"help"`
	ID   string     `yaml:"id"`
	Args []Argument `yaml:"args"`
}

type Argument struct {
	Name     string `yaml:"name"`
	Value    string `yaml:"default"`
	Optional bool   `yaml:"optional"`
}

func Actions(a ActionFile) []Action {
	return a.Actions
}

func List(file string) ActionFile {
	yfile, err := ioutil.ReadFile(file)

	if err != nil {

		log.Fatal(err)
	}

	data := ActionFile{}

	err2 := yaml.Unmarshal(yfile, &data)

	if err2 != nil {

		log.Fatal(err2)
	}

	return data
}

func Confirm(act Action) (bool, error) {
	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("Run %v", act.Name),
		IsConfirm: true,
		Default:   "n",
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
