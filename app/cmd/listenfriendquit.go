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
	"github.com/kprc/chatclient/app/cmdclient"
	"github.com/kprc/chatclient/app/cmdcommon"
	"log"

	"github.com/spf13/cobra"
)

var listenquitfriendaddr string

// listenfriendquitCmd represents the listenfriendquit command
var listenfriendquitCmd = &cobra.Command{
	Use:   "friend",
	Short: "quit friend listen service",
	Long:  `quit friend listen service`,
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := cmdcommon.IsProcessStarted(); err != nil {
			log.Println(err)
			return
		}

		if listenquitfriendaddr == "" {
			log.Println("please input friend address")
			return
		}

		var param []string
		param = append(param, listenquitfriendaddr)

		cmdclient.StringOpCmdSend("", cmdcommon.CMD_QUIT_LISTEN_FRIEND, param)
	},
}

func init() {
	listenQuitCmd.AddCommand(listenfriendquitCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listenfriendquitCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listenfriendquitCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	listenfriendquitCmd.Flags().StringVarP(&listenquitfriendaddr, "friend", "f", "", "quit listen friend address")
}
