/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/jatalocks/opsilon/pkg/run"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run an available workflow",
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		run.Select(repoNameRun, workflowName, inputs, confirm)
	},
}

var repoNameRun string
var workflowName string
var inputs map[string]string
var confirm bool

func init() {
	rootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")
	runCmd.Flags().StringVarP(&repoNameRun, "reop", "r", "", "Repository Name")
	runCmd.Flags().StringVarP(&workflowName, "workflow", "w", "", "ID of the workflow to run")
	runCmd.Flags().BoolVar(&confirm, "confirm", false, "Start running without confirmation")
	runCmd.Flags().StringToStringVarP(&inputs, "args", "a", nil, "Comma separated list of key=value arguments for the workflow input")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
