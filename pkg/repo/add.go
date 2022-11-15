package repo

import (
	"errors"
	"fmt"

	"github.com/jatalocks/opsilon/internal/config"
	"github.com/manifoldco/promptui"
)

func Add() {
	repoList := config.GetRepoList()
	fmt.Println("Current Repositories", repoList)
	validate := func(input string) error {
		if config.StringInSlice(input, repoList) {
			return errors.New("repo with the name already exists in your configuration file")
		} else if input == "" {
			return errors.New("repo cannot be blank.")
		} else {
			return nil

		}
	}

	prompt := promptui.Prompt{
		Label:    "Repository",
		Validate: validate,
	}
	result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	fmt.Printf("You choose %q\n", result)
}
