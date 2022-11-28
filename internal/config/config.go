package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/jatalocks/opsilon/internal/engine"
	"github.com/jatalocks/opsilon/internal/internaltypes"
	"github.com/jatalocks/opsilon/internal/logger"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

type Location struct {
	Path      string `json:"path" xml:"path" form:"path" query:"path" mapstructure:"path" validate:"nonzero"`
	Type      string `json:"type" xml:"type" form:"type" query:"type" mapstructure:"type" validate:"nonzero"`
	Subfolder string `json:"subfolder" xml:"subfolder" form:"subfolder" query:"subfolder" mapstructure:"subfolder,omitempty"`
	Branch    string `json:"branch" xml:"branch" form:"branch" query:"branch" mapstructure:"branch,omitempty"`
}

type Repo struct {
	Name        string   `json:"name" xml:"name" form:"name" query:"name" mapstructure:"name" validate:"nonzero"`
	Description string   `json:"description" xml:"description" form:"description" query:"description" mapstructure:"description"`
	Location    Location `json:"location" xml:"location" form:"location" query:"location" mapstructure:"location" validate:"nonzero"`
}

type RepoFile struct {
	Repositories []Repo `mapstructure:"repositories" validate:"nonzero"`
}

var C RepoFile

func PrintRepos(repos []Repo) {
	var data [][]string

	for _, r := range repos {
		row := []string{r.Name, r.Description, r.Location.Path, r.Location.Type, r.Location.Branch, r.Location.Subfolder}
		data = append(data, row)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Description", "Path/URL", "Type", "Branch", "Subfolder"})

	for _, v := range data {
		table.Append(v)
	}
	table.Render() // Send output
}

func countRune(s string, r rune) int {
	count := 0
	for _, c := range s {
		if c == r {
			count++
		}
	}
	return count
}
func PrintStageResults(results []internaltypes.Result) {
	var data [][]string

	for _, r := range results {
		// for _, v := range r.Logs {
		// 	fmt.Println(r.Stage.ID, v)
		// }
		row := []string{r.Stage.Stage, r.Stage.ID, fmt.Sprint(r.Result), fmt.Sprint(r.Skipped), fmt.Sprint(engine.GenEnv(r.Outputs)), fmt.Sprint(len(r.Logs))}
		data = append(data, row)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Stage", "ID", "Result", "Skipped", "Outputs", "Log Lines"})

	for _, v := range data {
		table.Append(v)
	}
	table.Render() // Send output
}

func TrimSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}
	return s
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func PrintWorkflows(workflows []internaltypes.Workflow) {
	var data [][]string

	for _, r := range workflows {
		out := ""
		for _, v := range r.Input {
			out += fmt.Sprintf("%v,", v.Name)
		}
		images := []string{r.Image}
		for _, v := range r.Stages {
			if !StringInSlice(v.Image, images) {
				images = append(images, v.Image)
			}
		}

		row := []string{r.Repo, r.ID, r.Description, TrimSuffix(strings.Join(images, ","), ","), TrimSuffix(out, ","), strconv.Itoa(len(r.Stages))}
		data = append(data, row)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Repository", "ID", "Description", "Images Used", "Inputs", "Stage Count"})

	for _, v := range data {
		table.Append(v)
	}
	table.Render() // Send output
}

func GetConfig() []Repo {
	err2 := viper.Unmarshal(&C)
	logger.HandleErr(err2)
	return C.Repositories
}

func GetConfigFile() *RepoFile {
	err2 := viper.Unmarshal(&C)
	logger.HandleErr(err2)
	return &C
}

func SaveToConfig(r RepoFile) {
	if viper.GetBool("consul") {
		client, err := api.NewClient(api.DefaultConfig())
		if err != nil {
			panic(err)
		}

		// Get a handle to the KV API
		kv := client.KV()
		out, err := yaml.Marshal(r)
		logger.HandleErr(err)
		// PUT a new KV pair
		p := &api.KVPair{Key: "OPSILON", Value: out}
		_, err = kv.Put(p, nil)
		logger.HandleErr(err)
	} else {
		file, err := os.OpenFile(viper.ConfigFileUsed(), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
		if err != nil {
			log.Fatalf("error opening/creating file: %v", err)
		}
		defer file.Close()

		enc := yaml.NewEncoder(file)

		err = enc.Encode(r)
		if err != nil {
			log.Fatalf("error encoding: %v", err)
		}
	}
}

func GetRepoList() []string {
	temp := []string{}
	err2 := viper.Unmarshal(&C)
	logger.HandleErr(err2)
	for _, r := range C.Repositories {
		temp = append(temp, r.Name)
	}
	return temp
}
