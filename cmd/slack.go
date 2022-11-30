/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/jatalocks/opsilon/pkg/slack"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// slackCmd represents the slack command
var slackCmd = &cobra.Command{
	Use:   "slack",
	Short: "Runs opsilon as a socket-mode slack bot",
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		slack.App(viper.GetString("slack_bot_token"), viper.GetString("slack_app_token"))
	},
}
var slack_bot_token string
var slack_app_token string

func init() {
	rootCmd.AddCommand(slackCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// slackCmd.PersistentFlags().String("foo", "", "A help for foo")
	slackCmd.Flags().StringVarP(&slack_bot_token, "slack_bot_token", "b", "xoxb-123", "slack bot token")
	slackCmd.Flags().StringVarP(&slack_app_token, "slack_app_token", "a", "xapp-123", "slack app token")
	viper.BindPFlag("slack_bot_token", slackCmd.Flags().Lookup("slack_bot_token"))
	viper.BindPFlag("slack_app_token", slackCmd.Flags().Lookup("slack_app_token"))
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// slackCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
