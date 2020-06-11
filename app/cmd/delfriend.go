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
	"fmt"

	"github.com/kprc/chatclient/app/cmdclient"
	"github.com/kprc/chatclient/app/cmdcommon"
	"github.com/spf13/cobra"
)

// delCmd represents the del command
var friendDelCmd = &cobra.Command{
	Use:   "del",
	Short: "del a friend",
	Long:  `del a friend`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			fmt.Println("command error")
			return
		}

		if friendChatAddr == "" {
			fmt.Println("need friend chat address")
			return
		}

		var param []string
		param = append(param, friendChatAddr)

		cmdclient.StringOpCmdSend("", cmdcommon.CMD_DEL_FRIEND, param)
	},
}

func init() {
	friendCmd.AddCommand(friendDelCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// delCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// delCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	friendDelCmd.Flags().StringVarP(&friendChatAddr, "address", "a", "", "chat address")
}
