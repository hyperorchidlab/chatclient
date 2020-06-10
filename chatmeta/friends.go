package chatmeta

import (
	"encoding/json"
	"github.com/btcsuite/btcutil/base58"
	"github.com/kprc/chat-protocol/address"
	"github.com/kprc/chat-protocol/groupid"
	"github.com/kprc/chat-protocol/protocol"
	"github.com/kprc/chatclient/chatcrypt"
	"github.com/kprc/chatclient/config"
	"github.com/kprc/chatclient/db"
	"github.com/kprc/nbsnetwork/tools"
	"github.com/pkg/errors"
	"github.com/rickeyliao/ServiceAgent/common"
	"log"
	"strconv"
)

func AddFriend(addr address.ChatAddress) error {
	if !addr.IsValid() {
		return errors.New("addr is error")
	}

	cfg := config.GetCCC()

	localaddr := address.ToAddress(cfg.PubKey)

	if localaddr == addr {
		return errors.New("don't add self as friend")
	}

	fd := &protocol.FriendDesc{}
	fd.PeerPubKey = addr.String()
	fd.SendTime = tools.GetNowMsTime()

	cmd := protocol.UserCommand{}
	cmd.Op = protocol.AddFriend
	cmd.SP = *cfg.SP

	serverPub := address.ChatAddress(cfg.SP.SignText.SPubKey).ToPubKey()
	aesk, _ := chatcrypt.GenerateAesKey(serverPub, cfg.PrivKey)

	data, _ := json.Marshal(*fd)

	ciphertxt, _ := chatcrypt.Encrypt(aesk, data)

	cmd.CipherTxt = base58.Encode(ciphertxt)

	d2s, _ := json.Marshal(cmd)

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

	if reply.CipherTxt != cmd.CipherTxt {
		return errors.New("add friend failed, cipher text is not equal")
	}

	if reply.ResultCode == 0 || reply.OP == protocol.AddFriend {
		return nil
	}

	return errors.New("add friend failed")

}

func ListGroupMembers(gid groupid.GrpID) error {
	cfg := config.GetCCC()

	cmd := protocol.UserCommand{}
	cmd.Op = protocol.ListGroupMbr
	cmd.SP = *cfg.SP

	var (
		key       []byte
		ciphertxt []byte
		err       error
		resp      string
		stat      int
	)

	key, err = chatcrypt.GenerateAesKey(address.ChatAddress(cfg.SP.SignText.SPubKey).ToPubKey(), cfg.PrivKey)
	if err != nil {
		return err
	}

	lgm := &protocol.ListGrpMbr{}
	lgm.GroupId = gid.String()

	plaintxt, _ := json.Marshal(*lgm)

	ciphertxt, err = chatcrypt.Encrypt(key, plaintxt)
	if err != nil {
		return err
	}

	cmd.CipherTxt = base58.Encode(ciphertxt)

	d2s, _ := json.Marshal(cmd)

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

	if reply.ResultCode != 0 {
		return errors.New("result code error")
	}

	plaintxt, err = chatcrypt.Decrypt(key, base58.Decode(reply.CipherTxt))
	if err != nil {
		return err
	}

	gml := &protocol.GroupMbrDetailsList{}

	err = json.Unmarshal(plaintxt, gml)
	if err != nil {
		return err
	}

	for i := 0; i < len(gml.FD); i++ {
		fd := &gml.FD[i]

		insert2GroupMember(gid, fd)
	}

	return nil

}

func ListFriends() (string, error) {

	cfg := config.GetCCC()

	cmd := protocol.UserCommand{}
	cmd.Op = protocol.ListFriend
	cmd.SP = *cfg.SP

	d2s, _ := json.Marshal(cmd)

	var (
		resp string
		stat int
		err  error
	)
	log.Println(string(d2s))

	resp, stat, err = common.Post1(config.GetCCC().GetCmdUrl(), string(d2s), false)
	if err != nil {
		return "", err
	}
	if stat != 200 {
		return "", errors.New("Get Error Stat Code:" + strconv.Itoa(stat))
	}

	log.Println(resp)

	reply := &protocol.UCReply{}
	json.Unmarshal([]byte(resp), reply)

	var (
		key, plaintxt []byte
	)

	key, err = chatcrypt.GenerateAesKey(address.ChatAddress(cfg.SP.SignText.SPubKey).ToPubKey(), cfg.PrivKey)
	if err != nil {
		return "", err
	}

	plaintxt, err = chatcrypt.Decrypt(key, base58.Decode(reply.CipherTxt))
	if err != nil {
		return "", err
	}

	fl := &protocol.FriendList{}

	err = json.Unmarshal(plaintxt, fl)
	if err != nil {
		return "", err
	}

	//insert into db
	insert2FriendDB(fl)

	for i := 0; i < len(fl.GD); i++ {
		g := fl.GD[i]

		err = ListGroupMembers(groupid.GrpID(g.GrpId))
		if err != nil {
			log.Println(err)
		}
	}

	return "refresh success", nil

}

func insert2GroupMember(gid groupid.GrpID, gm *protocol.GMember) {
	mdb := db.GetMetaDb()

	mdb.AddGroupMember(gid, gm.Alias, address.ChatAddress(gm.PubKey), gm.Agree, gm.ExpireTime)
}

func insert2FriendDB(fl *protocol.FriendList) {

	mdb := db.GetMetaDb()

	for i := 0; i < len(fl.FD); i++ {
		f := fl.FD[i]
		mdb.AddFriend(f.Alias, address.ChatAddress(f.PubKey), f.Agree, f.AddTime)
	}

	for i := 0; i < len(fl.GD); i++ {
		g := fl.GD[i]
		mdb.AddGroup(g.Alias, groupid.GrpID(g.GrpId), g.IsOwner, g.CreateTime)
	}

}
