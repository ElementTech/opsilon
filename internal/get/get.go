package get

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

type ActionFile struct {
	Actions []Action `yaml:"actions"`
}

type Action struct {
	Name string     `yaml:"name"`
	Help string     `yaml:"help"`
	Args []Argument `yaml:"args"`
}

type Argument struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"`
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
