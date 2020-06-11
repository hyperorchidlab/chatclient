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

	"github.com/kprc/chat-protocol/groupid"
	"github.com/kprc/chatclient/app/cmdclient"
	"github.com/kprc/chatclient/app/cmdcommon"
	"github.com/spf13/cobra"
)

var chatlistgroupid string

var groupmemberCmd = &cobra.Command{
	Use:   "groupmember",
	Short: "show group info and group member",
	Long:  `show group info and group member`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			fmt.Println("command error")
			return
		}

		if chatlistgroupid == "" {
			fmt.Println("must input group id")
			return
		}
		if !groupid.GrpID(chatlistgroupid).IsValid() {
			fmt.Println("please input a correc group id")
		}

		var param []string
		param = append(param, chatlistgroupid)

		cmdclient.StringOpCmdSend("", cmdcommon.CMD_LIST_GROUPMBRS, param)
	},
}

func init() {
	rootCmd.AddCommand(groupmemberCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// groupmemberCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// groupmemberCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	groupmemberCmd.Flags().StringVarP(&chatlistgroupid, "groupid", "g", "", "group id")
}
