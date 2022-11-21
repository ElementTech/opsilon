/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/jatalocks/opsilon/pkg/web"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var port int64

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Runs an api server that functions the same as the CLI",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		web.App(port)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	serverCmd.Flags().Int64VarP(&port, "port", "p", 8080, "Port to start the web server in")
	serverCmd.Flags().Bool("kubernetes", false, "Run in Kubernetes instead of Docker. You must be connected to a Kubernetes Context")
	viper.BindPFlag("kubernetes", serverCmd.Flags().Lookup("kubernetes"))
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
