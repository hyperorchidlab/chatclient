/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"github.com/kprc/chatclient/app/cmdcommon"
	"github.com/kprc/chatclient/config"

	"github.com/spf13/cobra"
	"log"
)

var remoteserver string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init chat client",
	Long:  `init chat client`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		_, err = cmdcommon.IsProcessCanStarted()
		if err != nil {
			log.Println(err)
			return
		}

		InitCfg()

		cfg := config.GetCCC()

		cfg.Save()

	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")
	//initCmd.Flags().StringVarP(&keypassword, "password", "p", "", "password for key encrypt")
	initCmd.Flags().StringVarP(&remoteserver, "host", "r", "", "remote server ip address")
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
