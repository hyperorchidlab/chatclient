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
	"github.com/hyperorchidlab/chatclient/app/cmdclient"
	"github.com/hyperorchidlab/chatclient/app/cmdcommon"
	"github.com/spf13/cobra"
)

var groupchatname string

// createCmd represents the create command
var groupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "create a group",
	Long:  `create a group`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			fmt.Println("Command error")
			return
		}
		if groupchatname == "" {
			fmt.Println("need set a group name")
			return
		}
		var param []string
		param = append(param, groupchatname)

		cmdclient.StringOpCmdSend("", cmdcommon.CMD_CREATE_GROUP, param)

	},
}

func init() {
	groupCmd.AddCommand(groupCreateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	groupCreateCmd.Flags().StringVarP(&groupchatname, "name", "n", "", "group name")

}
