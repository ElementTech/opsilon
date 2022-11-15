/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/jatalocks/opsilon/pkg/repo"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a repo from your local config.",
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		repo.Delete(repoList)
	},
}

func init() {
	repoCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().StringSliceVarP(&repoList, "repo", "r", nil, "Comma seperated list of repositories to fetch workflows from.")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deleteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deleteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
