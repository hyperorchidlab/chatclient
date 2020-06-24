package api

import (
	"context"
	"github.com/kprc/chatclient/app/cmdcommon"
	"github.com/kprc/chatclient/app/cmdpb"

	"fmt"
	"github.com/kprc/chat-protocol/address"
	"github.com/kprc/chat-protocol/groupid"
	"github.com/kprc/chatclient/chatcrypt"
	"github.com/kprc/chatclient/chatmeta"
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
	case cmdcommon.CMD_ADD_FRIEND:
		if len(so.Param) != 1 {
			msg = "Param error"
		} else {
			msg = addFriend(so.Param[0])
		}
	case cmdcommon.CMD_DEL_FRIEND:
		if len(so.Param) != 1 {
			msg = "Param error"
		} else {
			msg = delFriend(so.Param[0])
		}

	case cmdcommon.CMD_CREATE_GROUP:
		if len(so.Param) != 1 {
			msg = "Param error"
		} else {
			msg = createGroup(so.Param[0])
		}
	case cmdcommon.CMD_DEL_GROUP:
		if len(so.Param) != 1 {
			msg = "Param error"
		} else {
			msg = delGroup(so.Param[0])
		}

	case cmdcommon.CMD_JOIN_GROUP:
		if len(so.Param) != 2 {
			msg = "Param error"
		} else {
			msg = joinGroup(so.Param[0], so.Param[1])
		}
	case cmdcommon.CMD_QUIT_GROUP:
		if len(so.Param) != 2 {
			msg = "Param error"
		} else {
			msg = quitGroup(so.Param[0], so.Param[1])
		}
	case cmdcommon.CMD_LIST_GROUPMBRS:
		if len(so.Param) != 1 {
			msg = "Param error"
		} else {

			msg = cso.ListGroupMembers(so.Param[0])

			if msg == "" {
				msg = "no results"
			}
		}
	case cmdcommon.CMD_LISTEN_FRIEND:
		if len(so.Param) != 1 {
			msg = "Param error"
		} else {
			//c:=time.Tick(time.Second)
			//select {
			//case <-c:
			//
			//}
		}

	default:
		return encapResp("Command Not Found"), nil
	}

	return encapResp(msg), nil
}

func (cso *CmdStringOPSrv) ListGroupMembers(gid string) string {
	cfg := config.GetCCC()

	if cfg.SP == nil {
		return "Please Register first"
	}
	msg, err := chatmeta.ListGroupMembers(groupid.GrpID(gid))
	if err != nil {
		return err.Error()
	}

	return msg
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
	cfg := config.GetCCC()
	if cfg.PrivKey == nil {
		return "Please load account first"
	}

	tv, err := strconv.Atoi(timeInterval)
	if err != nil {
		return err.Error()
	}

	if err = chatmeta.RegChat(alias, tv); err != nil {
		return err.Error()
	}

	msg := "Registered success"
	msg += fmt.Sprintf("Name:%-30s ExpireTime:%-30s",
		cfg.SP.SignText.AliasName,
		int64time2string(cfg.SP.SignText.ExpireTime))

	return msg
}

func addFriend(addr string) string {
	cfg := config.GetCCC()
	if cfg.SP == nil {
		return "Please register first"
	}

	if err := chatmeta.AddFriend(address.ChatAddress(addr)); err != nil {
		return err.Error()
	}

	return "Add " + addr + " friend success"
}

func delFriend(addr string) string {
	cfg := config.GetCCC()
	if cfg.SP == nil {
		return "Please register first"
	}

	if err := chatmeta.DelFriend(address.ChatAddress(addr)); err != nil {
		return err.Error()
	}

	return "Del " + addr + " friend success"
}

func createGroup(name string) string {
	cfg := config.GetCCC()

	if cfg.SP == nil {
		return "Please register first"
	}

	if err := chatmeta.CreateGroup(name); err != nil {
		return err.Error()
	}

	return "Create group " + name + " success"
}

func delGroup(gid string) string {
	cfg := config.GetCCC()

	if cfg.SP == nil {
		return "Please register first"
	}
	if !groupid.GrpID(gid).IsValid() {
		return "not a group id"
	}

	if err := chatmeta.DelGroup(groupid.GrpID(gid)); err != nil {
		return err.Error()
	}

	return "Delete group " + gid + " success"
}

func joinGroup(groupId string, userid string) string {
	cfg := config.GetCCC()
	if cfg.SP == nil {
		return "Please register first"
	}

	if err := chatmeta.JoinGroup(groupid.GrpID(groupId), userid); err != nil {
		return err.Error()
	}

	return "Join group success"

}

func quitGroup(groupId string, userid string) string {
	cfg := config.GetCCC()
	if cfg.SP == nil {
		return "Please register first"
	}

	if err := chatmeta.QuitGroup(groupid.GrpID(groupId), userid); err != nil {
		return err.Error()
	}

	return "Quit group success"

}

func int64time2string(t int64) string {
	tm := time.Unix(t/1000, 0)
	return tm.Format("2006-01-02 15:04:05")
}
