package chatmeta

import (
	"encoding/json"
	"errors"
	"github.com/hyperorchidlab/chat-protocol/address"
	"github.com/hyperorchidlab/chat-protocol/protocol"
	"github.com/hyperorchidlab/chatclient/config"
	"github.com/hyperorchidlab/chatserver/app/cmdcommon"
	"log"
)

func RegChat(alias string, months int) error {

	if alias == "" || months < 1 || months > 36 {
		return errors.New("param error")
	}

	urr := &protocol.UserRegReq{}

	urr.CPubKey = address.ToAddress(config.GetCCC().PubKey).String()
	urr.AliasName = alias
	urr.TimeInterval = int64(months)

	regs, err := json.Marshal(*urr)
	if err != nil {
		return err
	}

	var (
		resp string
		stat int
	)
	log.Println(string(regs))

	resp, stat, err = cmdcommon.Post1(config.GetCCC().GetRegUrl(), string(regs), false)
	if err != nil {
		return err
	}

	if stat != 200 {
		return errors.New("code is not 200")
	}

	log.Println(resp)

	uresp := &protocol.UserRegResp{}

	err = json.Unmarshal([]byte(resp), uresp)
	if err != nil {
		return err
	}

	if uresp.ErrCode != 0 {
		return errors.New("register error")
	}

	config.SaveUserIdentify(&uresp.SP)

	return nil
}
