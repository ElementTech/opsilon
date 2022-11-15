/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/jatalocks/opsilon/pkg/repo"
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var rlistCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available repositories",
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		repo.List()
	},
}

func init() {
	repoCmd.AddCommand(rlistCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
