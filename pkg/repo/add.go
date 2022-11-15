package repo

import (
	"errors"
	"fmt"

	"github.com/jatalocks/opsilon/internal/config"
	"github.com/jatalocks/opsilon/internal/get"
	"github.com/jatalocks/opsilon/internal/logger"
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
	PromptString(&repo.Name, "Repository", func(input string) error {
		if config.StringInSlice(input, repoList) {
			return errors.New("repo with the name already exists in your configuration file")
		} else if input == "" {
			return errors.New("repo cannot be blank")
		}
		return nil

	})
	PromptString(&repo.Description, "Description", func(input string) error { return nil })

	promptType := &promptui.Select{
		Label: "Select Repo Type",
		Items: []string{"folder"},
	}
	i, _, err := promptType.Run()
	logger.HandleErr(err)
	repo.Location.Type = []string{"folder"}[i]

	PromptString(&repo.Location.Path, "Path", func(input string) error { return nil })
	validate.ValidateRepo(&repo)
	fileConfig := config.GetConfigFile()
	fileConfig.Repositories = append(fileConfig.Repositories, repo)
	validate.ValidateRepoFile(fileConfig)
	viper.Set("", fileConfig)
	viper.WriteConfig()
	logger.HandleErr(err)

	config.SaveToConfig(*fileConfig)
	viper.ReadInConfig()
	_, err = get.GetWorkflowsForRepo([]string{repo.Name})
	if err != nil {
		Delete([]string{repo.Name})
		logger.HandleErr(err)
	}
	List()
}
