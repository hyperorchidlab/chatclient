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
	"strconv"
)

var regaliasname string
var regintervalmonth int32

// regCmd represents the reg command
var regCmd = &cobra.Command{
	Use:   "reg",
	Short: "register client",
	Long:  `register client`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) > 0 {
			fmt.Println("command error")
			return
		}

		if regaliasname == "" {
			fmt.Println("need alias param")
			return
		}

		if regintervalmonth <= 0 {
			fmt.Println("month interval value less than 36 and large than 0")
			return
		}

		var param []string
		param = append(param, regaliasname, strconv.Itoa(int(regintervalmonth)))

		cmdclient.StringOpCmdSend("", cmdcommon.CMD_REG_USER, param)
	},
}

func init() {
	rootCmd.AddCommand(regCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// regCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// regCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	regCmd.Flags().StringVarP(&regaliasname, "alias", "a", "", "user alias")
	regCmd.Flags().Int32VarP(&regintervalmonth, "month", "m", 0, "license to user month interval")
}
