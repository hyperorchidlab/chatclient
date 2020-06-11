package chatmeta

import (
	"encoding/json"
	"errors"
	"github.com/btcsuite/btcutil/base58"
	"github.com/kprc/chat-protocol/address"
	"github.com/kprc/chat-protocol/groupid"
	"github.com/kprc/chat-protocol/protocol"
	"github.com/kprc/chatclient/chatcrypt"
	"github.com/kprc/chatclient/config"
	"github.com/kprc/nbsnetwork/tools"
	"github.com/rickeyliao/ServiceAgent/common"
	"log"
	"strconv"
)

func CreateGroup(groupName string) error {

	gd := &protocol.GroupDesc{}
	gd.GroupAlias = groupName
	gd.GroupID = groupid.NewGroupId().String()
	gd.SendTime = tools.GetNowMsTime()

	cfg := config.GetCCC()

	uc := &protocol.UserCommand{}
	uc.Op = protocol.AddGroup
	uc.SP = *cfg.SP

	serverPub := address.ChatAddress(cfg.SP.SignText.SPubKey).ToPubKey()
	aesk, _ := chatcrypt.GenerateAesKey(serverPub, cfg.PrivKey)

	data, _ := json.Marshal(*gd)

	ciphertxt, _ := chatcrypt.Encrypt(aesk, data)

	uc.CipherTxt = base58.Encode(ciphertxt)

	d2s, _ := json.Marshal(uc)

	var (
		resp string
		stat int
		err  error
	)
	log.Println(string(d2s))

	resp, stat, err = common.Post1(config.GetCCC().GetCmdUrl(), string(d2s), false)
	if err != nil {
		return err
	}
	if stat != 200 {
		return errors.New("Get Error Stat Code:" + strconv.Itoa(stat))
	}

	log.Println(resp)

	reply := &protocol.UCReply{}
	json.Unmarshal([]byte(resp), reply)

	if reply.CipherTxt != uc.CipherTxt {
		return errors.New("create group failed, cipher text is not equal")
	}

	if reply.ResultCode == 0 || reply.OP == protocol.AddGroup {
		return nil
	}

	return errors.New("create group failed")

}

func DelGroup(groupName string) error {

	gd := &protocol.GroupDesc{}
	gd.GroupAlias = groupName
	gd.GroupID = groupid.NewGroupId().String()
	gd.SendTime = tools.GetNowMsTime()

	cfg := config.GetCCC()

	uc := &protocol.UserCommand{}
	uc.Op = protocol.DelGroup
	uc.SP = *cfg.SP

	serverPub := address.ChatAddress(cfg.SP.SignText.SPubKey).ToPubKey()
	aesk, _ := chatcrypt.GenerateAesKey(serverPub, cfg.PrivKey)

	data, _ := json.Marshal(*gd)

	ciphertxt, _ := chatcrypt.Encrypt(aesk, data)

	uc.CipherTxt = base58.Encode(ciphertxt)

	d2s, _ := json.Marshal(uc)

	var (
		resp string
		stat int
		err  error
	)
	log.Println(string(d2s))

	resp, stat, err = common.Post1(config.GetCCC().GetCmdUrl(), string(d2s), false)
	if err != nil {
		return err
	}
	if stat != 200 {
		return errors.New("Get Error Stat Code:" + strconv.Itoa(stat))
	}

	log.Println(resp)

	reply := &protocol.UCReply{}
	json.Unmarshal([]byte(resp), reply)

	if reply.CipherTxt != uc.CipherTxt {
		return errors.New("delete group failed, cipher text is not equal")
	}

	if reply.ResultCode == 0 || reply.OP == protocol.DelGroup {
		return nil
	}

	return errors.New("delete group failed")

}

func JoinGroup(gid groupid.GrpID, friendPk string) error {

	if !gid.IsValid() {
		return errors.New("group id is not corrected")
	}
	if !address.ChatAddress(friendPk).IsValid() {
		return errors.New("user id not correct")
	}

	cfg := config.GetCCC()

	uc := &protocol.UserCommand{}
	uc.Op = protocol.JoinGroup
	uc.SP = *cfg.SP

	gmd := &protocol.GroupMemberDesc{}
	gmd.GroupID = gid.String()
	gmd.Friend = friendPk
	gmd.SendTime = tools.GetNowMsTime()

	serverPub := address.ChatAddress(cfg.SP.SignText.SPubKey).ToPubKey()
	aesk, _ := chatcrypt.GenerateAesKey(serverPub, cfg.PrivKey)

	data, _ := json.Marshal(*gmd)

	ciphertxt, _ := chatcrypt.Encrypt(aesk, data)

	uc.CipherTxt = base58.Encode(ciphertxt)

	d2s, _ := json.Marshal(uc)

	var (
		resp string
		stat int
		err  error
	)
	log.Println(string(d2s))

	resp, stat, err = common.Post1(config.GetCCC().GetCmdUrl(), string(d2s), false)
	if err != nil {
		return err
	}
	if stat != 200 {
		return errors.New("Get Error Stat Code:" + strconv.Itoa(stat))
	}

	log.Println(resp)

	reply := &protocol.UCReply{}
	json.Unmarshal([]byte(resp), reply)

	if reply.CipherTxt != uc.CipherTxt {
		return errors.New("join group failed, cipher text is not equal")
	}

	if reply.ResultCode == 0 || reply.OP == protocol.AddGroup {
		return nil
	}

	return errors.New("join group failed")
}

func QuitGroup(gid groupid.GrpID, friendPk string) error {

	if !gid.IsValid() {
		return errors.New("group id is not corrected")
	}
	if !address.ChatAddress(friendPk).IsValid() {
		return errors.New("user id not correct")
	}

	cfg := config.GetCCC()

	uc := &protocol.UserCommand{}
	uc.Op = protocol.QuitGroup
	uc.SP = *cfg.SP

	gmd := &protocol.GroupMemberDesc{}
	gmd.GroupID = gid.String()
	gmd.Friend = friendPk
	gmd.SendTime = tools.GetNowMsTime()

	serverPub := address.ChatAddress(cfg.SP.SignText.SPubKey).ToPubKey()
	aesk, _ := chatcrypt.GenerateAesKey(serverPub, cfg.PrivKey)

	data, _ := json.Marshal(*gmd)

	ciphertxt, _ := chatcrypt.Encrypt(aesk, data)

	uc.CipherTxt = base58.Encode(ciphertxt)

	d2s, _ := json.Marshal(uc)

	var (
		resp string
		stat int
		err  error
	)
	log.Println(string(d2s))

	resp, stat, err = common.Post1(config.GetCCC().GetCmdUrl(), string(d2s), false)
	if err != nil {
		return err
	}
	if stat != 200 {
		return errors.New("Get Error Stat Code:" + strconv.Itoa(stat))
	}

	log.Println(resp)

	reply := &protocol.UCReply{}
	json.Unmarshal([]byte(resp), reply)

	if reply.CipherTxt != uc.CipherTxt {
		return errors.New("quit group failed, cipher text is not equal")
	}

	if reply.ResultCode == 0 || reply.OP == protocol.QuitGroup {
		return nil
	}

	return errors.New("quit group failed")
}
