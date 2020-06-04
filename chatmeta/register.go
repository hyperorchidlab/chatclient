package chatmeta

import (
	"encoding/json"
	"errors"
	"github.com/kprc/chat-protocol/address"
	"github.com/kprc/chat-protocol/protocol"
	"github.com/kprc/chatclient/config"
	"github.com/rickeyliao/ServiceAgent/common"
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

	resp, stat, err = common.Post1(config.GetCCC().GetRegUrl(), string(regs), false)
	if err != nil || stat != 200 {
		return err
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
