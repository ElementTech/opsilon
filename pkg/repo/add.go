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

func PromptString(str *string, label string, validate promptui.ValidateFunc, def string) {
	prompt := promptui.Prompt{
		Label:    label,
		Default:  def,
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

	}, "")
	PromptString(&repo.Description, "Description", func(input string) error { return nil }, "")
	options := []string{"git", "folder"}
	promptType := &promptui.Select{
		Label: "Select Repo Type",
		Items: options,
	}
	i, _, err := promptType.Run()
	logger.HandleErr(err)
	repo.Location.Type = options[i]
	switch repo.Location.Type {
	case "folder":
		PromptString(&repo.Location.Path, "Folder Path", func(input string) error {
			if input == "" {
				return errors.New("folder path cannot be blank")
			}
			return nil
		}, "")
	case "git":
		PromptString(&repo.Location.Path, "Git URL", func(input string) error {
			if input == "" {
				return errors.New("URL cannot be blank")
			}
			return nil

		}, "")
		PromptString(&repo.Location.Branch, "Branch (Optional)", func(input string) error { return nil }, "")
		PromptString(&repo.Location.Subfolder, "Subfolder (Optional)", func(input string) error { return nil }, "")
	}
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
		get.CheckIfError(err)
	}
	List()
}
