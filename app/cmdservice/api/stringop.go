package api

import (
	"context"
	"github.com/kprc/chatclient/app/cmdcommon"
	"github.com/kprc/chatclient/app/cmdpb"

	"github.com/kprc/chat-protocol/address"
	"github.com/kprc/chatclient/chatcrypt"
	"github.com/kprc/chatclient/config"
	"strconv"
	"time"
)

type CmdStringOPSrv struct {
}

func (cso *CmdStringOPSrv) StringOpDo(cxt context.Context, so *cmdpb.StringOP) (*cmdpb.DefaultResp, error) {
	msg := ""
	switch so.Op {
	case cmdcommon.CMD_ACCOUNT_CREATE:
		msg = createAccount(so.Param[0])
	case cmdcommon.CMD_ACCOUNT_LOAD:
		msg = loadAccount(so.Param[0])
	case cmdcommon.CMD_REG_USER:
		if len(so.Param) != 2 {
			msg = "Param error"
		} else {
			msg = regUser(so.Param[0], so.Param[1])
		}

	default:
		return encapResp("Command Not Found"), nil
	}

	return encapResp(msg), nil
}

func createAccount(passwd string) string {
	err := chatcrypt.GenEd25519KeyAndSave(passwd)
	if err != nil {
		return "create account failed"
	}

	chatcrypt.LoadKey(passwd)

	addr := address.ToAddress(config.GetCCC().PubKey).String()

	return "Address: " + addr
}

func loadAccount(passwd string) string {

	chatcrypt.LoadKey(passwd)

	addr := address.ToAddress(config.GetCCC().PubKey).String()

	return "load account success! \r\nAddress: " + addr
}

func regUser(alias string, timeInterval string) string {
	//tm :=time.Now()
	//
	//tv,_:=strconv.Atoi(timeInterval)
	//
	//tm1:=tm.AddDate(0, tv,0)
	//
	//
	//
	//return tm1.Format("2006-01-02 15:04:05")
}

func int64time2string(t int64) string {
	tm := time.Unix(t/1000, 0)
	return tm.Format("2006-01-02 15:04:05")
}
