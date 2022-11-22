package repo

import (
	"github.com/jatalocks/opsilon/internal/config"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"
)

func RemoveIndex(s []config.Repo, index int) []config.Repo {
	return append(s[:index], s[index+1:]...)
}

func Delete(repoList []string) error {
	configFile := config.GetConfigFile()
	removeList := repoList
	currentList := config.GetRepoList()
	if len(repoList) == 0 {
		promptRepo := &promptui.Select{
			Label: "Select Repo",
			Items: currentList,
		}
		i, _, err := promptRepo.Run()
		removeList = append(removeList, currentList[i])
		if err != nil {
			return err
		}
	}
	for _, v := range removeList {
		configFile.Repositories = RemoveIndex(configFile.Repositories, slices.IndexFunc(configFile.Repositories, func(c config.Repo) bool { return c.Name == v }))
	}

	viper.Set("", configFile)
	err := viper.WriteConfig()
	if err != nil {
		return err
	}
	config.SaveToConfig(*configFile)
	err = viper.ReadInConfig()
	if err != nil {
		return err
	}
	List()
	return nil
}
