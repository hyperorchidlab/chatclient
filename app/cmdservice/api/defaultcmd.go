package api

import (
	"context"

	"time"

	"encoding/json"

	"github.com/kprc/chatclient/config"

	"github.com/kprc/chat-protocol/address"
	"github.com/kprc/chatclient/app/cmdcommon"
	"github.com/kprc/chatclient/app/cmdpb"

)

type CmdDefaultServer struct {
	Stop func()
}

func (cds *CmdDefaultServer) DefaultCmdDo(ctx context.Context,
	request *cmdpb.DefaultRequest) (*cmdpb.DefaultResp, error) {
	if request.Reqid == cmdcommon.CMD_STOP {
		return cds.stop()
	}

	if request.Reqid == cmdcommon.CMD_CONFIG_SHOW {
		return cds.configShow()
	}

	if request.Reqid == cmdcommon.CMD_PK_SHOW {
		return cds.accountShow()
	}

	if request.Reqid == cmdcommon.CMD_RUN {
		return cds.serverRun()
	}

	resp := &cmdpb.DefaultResp{}

	resp.Message = "no cmd found"

	return resp, nil
}

func (cds *CmdDefaultServer) stop() (*cmdpb.DefaultResp, error) {

	go func() {
		time.Sleep(time.Second * 2)
		cds.Stop()
	}()
	resp := &cmdpb.DefaultResp{}
	resp.Message = "server stoped"
	return resp, nil
}

func encapResp(msg string) *cmdpb.DefaultResp {
	resp := &cmdpb.DefaultResp{}
	resp.Message = msg

	return resp
}

func (cds *CmdDefaultServer) configShow() (*cmdpb.DefaultResp, error) {
	cfg := config.GetCCC()

	bapc, err := json.MarshalIndent(*cfg, "", "\t")
	if err != nil {
		return encapResp("Internal error"), nil
	}

	return encapResp(string(bapc)), nil
}

func (cds *CmdDefaultServer) accountShow() (*cmdpb.DefaultResp, error) {
	cfg := config.GetCCC()

	msg := "please create account"

	if cfg.PubKey != nil {
		msg = address.ToAddress(cfg.PubKey).String()
	}

	return encapResp(msg), nil
}

func (cds *CmdDefaultServer) serverRun() (*cmdpb.DefaultResp, error) {
	if config.GetCCC().PubKey == nil || config.GetCCC().PrivKey == nil {
		return encapResp("chat server need account"), nil
	}

	//go httpservice.StartWebDaemon()

	return encapResp("Server running"), nil
}
