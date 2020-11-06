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
	"github.com/hyperorchidlab/chatclient/app/cmdclient"
	"github.com/hyperorchidlab/chatclient/app/cmdcommon"
	"log"

	"github.com/spf13/cobra"
)

var (
	sendP2pMsg       string
	sendP2pMsgFriend string
)

// sendfriendmessageCmd represents the sendfriendmessage command
var sendfriendmessageCmd = &cobra.Command{
	Use:   "message",
	Short: "send a message to a friend",
	Long:  `send a message to a friend`,
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := cmdcommon.IsProcessStarted(); err != nil {
			log.Println(err)
			return
		}

		if sendP2pMsg == "" {
			log.Println("please input message")
			return
		}

		if sendP2pMsgFriend == "" {
			log.Println("please input friend")
			return
		}

		var param []string
		param = append(param, sendP2pMsg, sendP2pMsgFriend)

		cmdclient.StringOpCmdSend("", cmdcommon.CMD_SEND_P2PMSG, param)
	},
}

func init() {
	friendCmd.AddCommand(sendfriendmessageCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// sendfriendmessageCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// sendfriendmessageCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	sendfriendmessageCmd.Flags().StringVarP(&sendP2pMsg, "message", "m", "", "message to send friend")
	sendfriendmessageCmd.Flags().StringVarP(&sendP2pMsgFriend, "friend", "f", "", "friend for receive the message")
}
