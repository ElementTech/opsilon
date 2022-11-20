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

func Add(name string, desc string, rtype string, path string, branch string, subfolder string) {
	repoList := config.GetRepoList()
	repo := config.Repo{}
	if name == "" {
		PromptString(&repo.Name, "Repository", func(input string) error {
			if config.StringInSlice(input, repoList) {
				return errors.New("repo with the name already exists in your configuration file")
			} else if input == "" {
				return errors.New("repo cannot be blank")
			}
			return nil
		}, "")
	} else {
		repo.Name = name
	}
	if desc == "" {
		PromptString(&repo.Description, "Description", func(input string) error { return nil }, "")
	} else {
		repo.Description = desc
	}
	if rtype == "" {
		options := []string{"git", "folder"}
		promptType := &promptui.Select{
			Label: "Select Repo Type",
			Items: options,
		}
		i, _, err := promptType.Run()
		logger.HandleErr(err)
		repo.Location.Type = options[i]
	} else {
		repo.Location.Type = rtype
	}
	switch repo.Location.Type {
	case "folder":
		if path == "" {
			PromptString(&repo.Location.Path, "Folder Path", func(input string) error {
				if input == "" {
					return errors.New("folder path cannot be blank")
				}
				return nil
			}, "")
		} else {
			repo.Location.Path = path
		}
	case "git":
		if path == "" {
			PromptString(&repo.Location.Path, "Git URL", func(input string) error {
				if input == "" {
					return errors.New("URL cannot be blank")
				}
				return nil
			}, "")
		} else {
			repo.Location.Path = path
		}
		if branch == "" {
			PromptString(&repo.Location.Branch, "Branch (Optional)", func(input string) error { return nil }, "")
		} else {
			repo.Location.Branch = branch
		}
		if subfolder == "" {
			PromptString(&repo.Location.Subfolder, "Subfolder (Optional)", func(input string) error { return nil }, "")
		} else {
			repo.Location.Subfolder = subfolder
		}
	}
	err := InsertRepositoryIfValid(repo)
	if err != nil {
		logger.HandleErr(err)
	}
	List()
}

func InsertRepositoryIfValid(repo config.Repo) error {
	err := validate.ValidateRepo(&repo)
	if err != nil {
		return err
	}
	fileConfig := config.GetConfigFile()
	fileConfig.Repositories = append(fileConfig.Repositories, repo)
	err = validate.ValidateRepoFile(fileConfig)
	if err != nil {
		return err
	}
	viper.Set("", fileConfig)
	err = viper.WriteConfig()
	if err != nil {
		return err
	}
	config.SaveToConfig(*fileConfig)
	err = viper.ReadInConfig()
	if err != nil {
		return err
	}
	_, err = get.GetWorkflowsForRepo([]string{repo.Name})
	if err != nil {
		Delete([]string{repo.Name})
		return err
	}
	return nil
}
