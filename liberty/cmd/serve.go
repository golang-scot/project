// Copyright © 2017 Graham Anderson <graham.anderson@gmail.com>
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
	"golang.scot/liberty"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "starts liberty proxy in server mode",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		serve()
	},
}

func init() {
	RootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	serveCmd.Flags().StringVarP(&srvAddr, "address", "a", "all", "IP address to bind to.")
}

var srvAddr string

func serve() {

	balancerConfig := &liberty.Config{
		Certs:     cfg.Certs,
		Proxies:   cfg.Proxies,
		Whitelist: cfg.Whitelist,
	}

	bl := liberty.NewProxy(balancerConfig)

	glog.Info("Router is bootstrapped, listening for connections...")
	bl.Serve()
}
