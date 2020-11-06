package api

import (
	"context"

	"time"

	"encoding/json"

	"github.com/hyperorchidlab/chatclient/config"

	"github.com/hyperorchidlab/chat-protocol/address"
	"github.com/hyperorchidlab/chatclient/app/cmdcommon"
	"github.com/hyperorchidlab/chatclient/app/cmdpb"
	"github.com/hyperorchidlab/chatclient/chatmeta"
	"github.com/hyperorchidlab/chatclient/db"
)

type CmdDefaultServer struct {
	Stop func()
}

func (cds *CmdDefaultServer) DefaultCmdDo(ctx context.Context,
	request *cmdpb.DefaultRequest) (*cmdpb.DefaultResp, error) {

	msg := ""

	switch request.Reqid {
	case cmdcommon.CMD_STOP:
		msg = cds.stop()
	case cmdcommon.CMD_CONFIG_SHOW:
		msg = cds.configShow()
	case cmdcommon.CMD_PK_SHOW:
		msg = cds.accountShow()
	case cmdcommon.CMD_RUN:
		msg = cds.serverRun()
	case cmdcommon.CMD_REFRESH_ALL:
		msg = cds.refreshAll()
	case cmdcommon.CMD_LIST_FRIEND:
		msg = cds.listFriends()
	case cmdcommon.CMD_LIST_GROUP:
		msg = cds.ListGroups()
	}

	if msg == "" {
		msg = "No Results"
	}

	resp := &cmdpb.DefaultResp{}
	resp.Message = msg

	return resp, nil

}

func (cds *CmdDefaultServer) stop() string {

	go func() {
		time.Sleep(time.Second * 2)
		cds.Stop()
	}()

	return "chat client stopped"
}

func encapResp(msg string) *cmdpb.DefaultResp {
	resp := &cmdpb.DefaultResp{}
	resp.Message = msg

	return resp
}

func (cds *CmdDefaultServer) configShow() string {
	cfg := config.GetCCC()

	bapc, err := json.MarshalIndent(*cfg, "", "\t")
	if err != nil {
		return "Internal error"
	}

	return string(bapc)
}

func (cds *CmdDefaultServer) accountShow() string {
	cfg := config.GetCCC()

	msg := "please create account"

	if cfg.PubKey != nil {
		msg = address.ToAddress(cfg.PubKey).String()
	}

	return msg
}

func (cds *CmdDefaultServer) serverRun() string {
	if config.GetCCC().PubKey == nil || config.GetCCC().PrivKey == nil {
		return "chat client need account"
	}

	return "chat client running"
}

func (cds *CmdDefaultServer) refreshAll() string {

	cfg := config.GetCCC()

	if cfg.SP == nil {
		return "Please Register first"
	}

	db.RenewMetaDb()

	msg, err := chatmeta.RefreshFriends()
	if err != nil {
		return err.Error()
	}

	return msg
}

func (cds *CmdDefaultServer) listFriends() string {
	cfg := config.GetCCC()
	if cfg.SP == nil {
		return "Please Register first"
	}

	msg, err := chatmeta.ListFriends()
	if err != nil {
		return err.Error()
	}

	return msg
}

func (cds *CmdDefaultServer) ListGroups() string {
	cfg := config.GetCCC()
	if cfg.SP == nil {
		return "Please Register first"
	}

	msg, err := chatmeta.ListGroups()
	if err != nil {
		return err.Error()
	}

	return msg
}
