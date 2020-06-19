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
	"github.com/kprc/chatclient/db"
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
	err = json.Unmarshal([]byte(resp), reply)
	if err != nil {
		return err
	}
	if reply.CipherTxt != "" {
		return errors.New("create group failed, cipher text is not equal")
	}

	if reply.ResultCode == 0 && reply.OP == protocol.AddGroup {
		cipherBytes := base58.Decode(reply.CipherTxt)
		var plaintxt []byte
		plaintxt, err = chatcrypt.Decrypt(aesk, cipherBytes)
		resp := &protocol.GroupResp{}
		err = json.Unmarshal(plaintxt, &resp.GCI)
		if err != nil {
			log.Println("group create info error")
			return nil
		}

		mdb := db.GetMetaDb()
		mdb.AddGroup(resp.GCI.GroupName, resp.GCI.GID, resp.GCI.IsOwner, resp.GCI.CreateTime)

		return nil
	}

	return errors.New("create group failed")

}

func DelGroup(gid groupid.GrpID) error {

	gd := &protocol.GroupDesc{}
	gd.GroupID = gid.String()
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

	if reply.ResultCode == 0 && reply.OP == protocol.DelGroup {
		mdb := db.GetMetaDb()
		mdb.DelGroup(gid)
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

	if reply.CipherTxt != "" {
		return errors.New("join group failed, cipher text is empty")
	}

	if reply.ResultCode == 0 && reply.OP == protocol.AddGroup {
		var plainTxt []byte
		plainTxt, err = chatcrypt.Decrypt(aesk, base58.Decode(reply.CipherTxt))
		if err != nil {
			log.Println(err)
			return nil
		}
		resp := &protocol.GroupMemberResp{}
		err = json.Unmarshal(plainTxt, &resp.GMAI)
		if err != nil {
			log.Println(err)
			return nil
		}

		mdb := db.GetMetaDb()
		gi := &resp.GMAI
		mdb.AddGroupMember(gi.GID, gi.FriendName, gi.FriendAddr, gi.Agree, gi.JoinTime)

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

	if reply.ResultCode == 0 && reply.OP == protocol.QuitGroup {
		mdb := db.GetMetaDb()
		mdb.DelGroupMember(gid, address.ChatAddress(friendPk))
		return nil
	}

	return errors.New("quit group failed")
}
