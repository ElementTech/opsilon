package repo

import (
	"errors"
	"fmt"

	"github.com/jatalocks/opsilon/internal/config"
	"github.com/jatalocks/opsilon/internal/validate"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

func PromptString(str *string, label string, validate promptui.ValidateFunc) {
	prompt := promptui.Prompt{
		Label:    label,
		Validate: validate,
	}
	result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	*str = result
}

func Add() {
	repoList := config.GetRepoList()
	repo := config.Repo{}
	fmt.Println(repoList)
	PromptString(&repo.Name, "Repository", func(input string) error {
		if config.StringInSlice(input, repoList) {
			return errors.New("repo with the name already exists in your configuration file")
		} else if input == "" {
			return errors.New("repo cannot be blank")
		}
		return nil

	})
	PromptString(&repo.Description, "Description", func(input string) error { return nil })
	PromptString(&repo.Location.Type, "Type", func(input string) error {
		if input != "folder" {
			return errors.New("invalid type. valid types are: folder")
		}
		return nil
	})
	PromptString(&repo.Location.Path, "Path", func(input string) error { return nil })
	validate.ValidateRepo(&repo)
	fileConfig := config.GetConfigFile()
	fileConfig.Repositories = append(fileConfig.Repositories, repo)
	validate.ValidateRepoFile(fileConfig)
	viper.Set("", fileConfig)
	viper.WriteConfig()

	config.SaveToConfig(*fileConfig)
	viper.ReadInConfig()
	List()
}
