/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/jatalocks/opsilon/pkg/run"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

var (
	repoNameRun  string
	workflowName string
	inputs       map[string]string
	confirm      bool
)

func init() {
	rootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	runCmd.Flags().Bool("kubernetes", false, "Run in Kubernetes instead of Docker. You must be connected to a Kubernetes Context")
	viper.BindPFlag("kubernetes", runCmd.Flags().Lookup("kubernetes"))
	runCmd.Flags().StringVarP(&repoNameRun, "repo", "r", "", "Repository Name")
	runCmd.Flags().StringVarP(&workflowName, "workflow", "w", "", "ID of the workflow to run")
	runCmd.Flags().BoolVar(&confirm, "confirm", false, "Start running without confirmation")
	runCmd.Flags().StringToStringVarP(&inputs, "args", "a", nil, "Comma separated list of key=value arguments for the workflow input")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
