package list

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
)

func List(file string) map[interface{}]interface{} {
	yfile, err := ioutil.ReadFile(file)

	if err != nil {

		log.Fatal(err)
	}

	data := make(map[interface{}]interface{})

	err2 := yaml.Unmarshal(yfile, &data)

	if err2 != nil {

		log.Fatal(err2)
	}

	return data
}
