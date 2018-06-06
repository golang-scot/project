// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"golang.scot/liberty"
	"golang.scot/liberty/middleware"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	configFile     = "liberty"
	configLocation = "/etc/liberty"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "liberty",
	Short: "Liberty is a reverse proxy with some basic load balancing features.",
	Long:  ``,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

}

// Config is the top level configuration for this package, at this moment the
// persisted paramaters are expected to be read from a yaml file.
type Config struct {
	Env           string                     `yaml:"env"`
	Profiling     bool                       `yaml:"profiling"`
	ProfStatsFile string                     `yaml:"profStatsFile"`
	Certs         []*liberty.Crt             `yaml:"certs"`
	Proxies       []*liberty.ReverseProxy    `yaml:"proxies"`
	Whitelist     []*middleware.ApiWhitelist `yaml:"whitelist"`
}

var cfg = &Config{}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName(configFile)
		viper.AddConfigPath(configLocation)
		viper.AddConfigPath("/liberty")
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("invlid config file: %s\n", viper.ConfigFileUsed())
		fmt.Println(err)
	}

	fmt.Println("Using config file:", viper.ConfigFileUsed())
	viper.AutomaticEnv() // read in environment variables that match
	viper.Unmarshal(cfg)
}
