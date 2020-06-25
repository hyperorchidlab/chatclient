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
	"github.com/kprc/chatclient/app/cmdlistenudp"
	"github.com/spf13/cobra"
	"log"
	"strconv"
)

var listenfriendaddr string

// listenfriendCmd represents the listenfriend command
var listenfriendCmd = &cobra.Command{
	Use:   "friend",
	Short: "listen a friend message",
	Long:  `listen a friend message`,
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := cmdcommon.IsProcessStarted(); err != nil {
			log.Println(err)
			return
		}

		if listenfriendaddr == "" {
			log.Println("please input friend address")
			return
		}

		port := cmdlistenudp.RandPort()

		var param []string
		param = append(param, listenfriendaddr, strconv.Itoa(port))

		server := cmdlistenudp.NewUdpServer(port)
		go server.Serve()

		cmdclient.StringOpCmdSend("", cmdcommon.CMD_LISTEN_FRIEND, param)

	},
}

func init() {
	listenCmd.AddCommand(listenfriendCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listenfriendCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listenfriendCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	listenfriendCmd.Flags().StringVarP(&listenfriendaddr, "friend", "f", "", "friend address")
}
