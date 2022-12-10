/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/hashicorp/consul/api"
	"github.com/jatalocks/opsilon/internal/config"
	"github.com/jatalocks/opsilon/internal/db"
	"github.com/jatalocks/opsilon/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "opsilon",
	Short: "A customizable CLI for collaboratively running container-native workflows",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { initConfig() },
}
var ver string

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string) {
	rootCmd.AddCommand(newVersionCmd(version)) // version subcommand
	ver = version
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// cobra.OnInitialize(initConfig)
	// rootCmd.InitDefaultHelpCmd()
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.opsilon.yaml)")
	rootCmd.PersistentFlags().Bool("kubernetes", false, "Run in Kubernetes instead of Docker. You must be connected to a Kubernetes Context")

	rootCmd.PersistentFlags().Bool("local", true, "Run using a local file as config. Not a database. True for CLI.")

	rootCmd.PersistentFlags().Bool("database", false, "Run using a MongoDB database.")

	rootCmd.PersistentFlags().Bool("consul", false, "Run using a Consul Key/Value store. This is for distributed installation.")

	rootCmd.MarkFlagsMutuallyExclusive("local", "database")
	rootCmd.PersistentFlags().String("mongodb_uri", "mongodb://localhost:27017", "Mongodb URI. Can be set using ENV variable.")

	rootCmd.PersistentFlags().String("consul_uri", "localhost:8500", "Consul URI. Can be set using ENV variable.")

	rootCmd.PersistentFlags().String("consul_key", "default", "Consul Config Key. Can be set using ENV variable.")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	consul, err := rootCmd.Flags().GetBool("consul")
	viper.Set("consul", consul)
	logger.HandleErr(err)
	if consul {
		consul_uri, err := rootCmd.Flags().GetString("consul_uri")
		viper.Set("consul_uri", consul_uri)
		logger.HandleErr(err)
		consul_key, err := rootCmd.Flags().GetString("consul_key")
		viper.Set("consul_key", consul_key)
		logger.HandleErr(err)
		viper.AddRemoteProvider("consul", consul_uri, consul_key)
		viper.SetConfigType("yaml") // Need to explicitly set this to yaml

		// Get a new client
		client, err := api.NewClient(&api.Config{
			Address: consul_uri,
		})
		logger.HandleErr(err)
		// Get a handle to the KV API
		kv := client.KV()
		// PUT a new KV pair
		err = viper.ReadRemoteConfig()
		if err != nil {
			p := &api.KVPair{Key: "default", Value: []byte("")}
			_, err = kv.Put(p, nil)
			logger.HandleErr(err)
		}
		viper.AutomaticEnv() // read in environment variables that match

	} else {
		if cfgFile != "" {
			// Use config file from the flag.
			viper.SetConfigFile(cfgFile)
		} else {
			// Find home directory.
			home, err := os.UserHomeDir()
			cobra.CheckErr(err)

			// Search config in home directory with name ".opsilon" (without extension).
			viper.AddConfigPath(home)
			viper.SetConfigType("yaml")
			viper.SetConfigName(".opsilon")
			viper.SafeWriteConfig()

		}
		viper.AutomaticEnv() // read in environment variables that match

		// If a config file is found, read it in.
		err = viper.ReadInConfig()
	}

	viper.BindPFlag("kubernetes", rootCmd.Flags().Lookup("kubernetes"))
	viper.BindPFlag("local", rootCmd.Flags().Lookup("local"))
	viper.BindPFlag("database", rootCmd.Flags().Lookup("database"))
	viper.BindPFlag("consul", rootCmd.Flags().Lookup("consul"))
	viper.BindPFlag("mongodb_uri", rootCmd.Flags().Lookup("mongodb_uri"))
	viper.BindPFlag("consul_uri", rootCmd.Flags().Lookup("consul_uri"))
	viper.BindPFlag("consul_key", rootCmd.Flags().Lookup("consul_key"))

	if err != nil {
		logger.Error("It seems that you don't yet have a repository config file. Please run:")
		logger.Info("opsilon repo add")
	}
	err2 := viper.Unmarshal(&[]config.RepoFile{})
	if err2 != nil {
		logger.Error("There appears to be a problem with your configuration. Please refer to the docs or run opsilon repo command.")
	}
	logger.HandleErr(err2)
	consul_uri, err := rootCmd.Flags().GetString("consul_uri")
	logger.HandleErr(err)
	if err == nil {
		if consul {
			fmt.Fprintln(os.Stderr, "Using consul", consul_uri)
		} else {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}

	db.Init()
}
