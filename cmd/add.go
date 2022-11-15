/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/jatalocks/opsilon/internal/config"
	"github.com/jatalocks/opsilon/pkg/repo"
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a workflow repo",
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		rtype := ""
		if folderType {
			rtype = "folder"
		}
		if gitType {
			rtype = "git"
		}
		repo.Add(repoName, repoDesc, rtype, location.Path, location.Branch, location.Subfolder)
	},
}

var repoName string
var repoDesc string
var location config.Location
var folderType bool
var gitType bool

func init() {
	repoCmd.AddCommand(addCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")
	addCmd.Flags().StringVarP(&repoName, "name", "n", "", "Repository Name")
	addCmd.Flags().StringVarP(&repoDesc, "desc", "d", "", "Description of the repository")
	addCmd.Flags().BoolVar(&folderType, "folder", false, "Repo of type Folder")
	addCmd.Flags().BoolVar(&gitType, "git", false, "Repo of type Git")
	addCmd.MarkFlagsMutuallyExclusive("folder", "git")
	addCmd.Flags().StringVarP(&location.Path, "path", "p", "", "Path/URL")
	addCmd.Flags().StringVarP(&location.Subfolder, "subfolder", "s", "", "Subfolder Path (git only)")
	addCmd.Flags().StringVarP(&location.Branch, "branch", "b", "", "Branch Name (git only)")
	addCmd.MarkFlagsMutuallyExclusive("folder", "subfolder")
	addCmd.MarkFlagsMutuallyExclusive("folder", "branch")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
