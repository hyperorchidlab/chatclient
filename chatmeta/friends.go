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
	err = json.Unmarshal([]byte(resp), reply)
	if err != nil {
		return err
	}

	if reply.CipherTxt == "" {
		return errors.New("add friend failed, cipher text is not set")
	}

	if reply.ResultCode == 0 && reply.OP == protocol.AddFriend {

		cipherbytes := base58.Decode(reply.CipherTxt)
		var plaintxt []byte
		plaintxt, err = chatcrypt.Decrypt(aesk, cipherbytes)
		resp := &protocol.FriendAddResp{}
		err = json.Unmarshal(plaintxt, &resp.FAI)
		if err != nil {
			log.Println("friend info error")
			return nil
		}

		mdb := db.GetMetaDb()
		mdb.AddFriend(resp.FAI.AliasName, resp.FAI.Addr, resp.FAI.Agree, resp.FAI.AddTime)

		return nil
	}

	return errors.New("add friend failed")

}

func DelFriend(addr address.ChatAddress) error {
	if !addr.IsValid() {
		return errors.New("addr is error")
	}

	cfg := config.GetCCC()

	localaddr := address.ToAddress(cfg.PubKey)

	if localaddr == addr {
		return errors.New("can't delete friend")
	}

	fd := &protocol.FriendDesc{}
	fd.PeerPubKey = addr.String()
	fd.SendTime = tools.GetNowMsTime()

	cmd := protocol.UserCommand{}
	cmd.Op = protocol.DelFriend
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
		return errors.New("del friend failed, cipher text is not equal")
	}

	if reply.ResultCode == 0 && reply.OP == protocol.DelFriend {
		mdb := db.GetMetaDb()
		mdb.DelFriend(addr)
		return nil
	}

	return errors.New("del friend failed")

}

func RefreshGroupMembers(gid groupid.GrpID) error {
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

	db.UpdateGroupKeyMem(gid, gml.Hashk)
	gkdb := db.GetChatGrpKeysDb()
	gkdb.Insert2(gml.Gkeys, gml.PKeys)

	for i := 0; i < len(gml.FD); i++ {
		fd := &gml.FD[i]
		insert2GroupMember(gid, fd)
	}

	return nil

}

func RefreshFriends() (string, error) {

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

		err = RefreshGroupMembers(groupid.GrpID(g.GrpId))
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

func ListFriends() (string, error) {
	mdb := db.GetMetaDb()

	type friendArg struct {
		msg string
	}

	arg := &friendArg{}

	mdb.TraversFriends(arg, func(arg, v interface{}) (ret interface{}, err error) {
		f := v.(*db.Friend)
		ag := arg.(*friendArg)

		if ag.msg != "" {
			ag.msg += "\r\n"
		}
		ag.msg += f.String()

		return nil, nil
	})

	return arg.msg, nil
}

func ListGroups() (string, error) {
	mdb := db.GetMetaDb()

	type groupArg struct {
		msg string
	}

	arg := &groupArg{}

	mdb.TraversGroups(arg, func(arg, v interface{}) (ret interface{}, err error) {
		g := v.(*db.Group)
		ag := arg.(*groupArg)

		if ag.msg != "" {
			ag.msg += "\r\n"
		}
		ag.msg += g.String()

		return nil, nil
	})

	return arg.msg, nil
}

func ListGroupMembers(gid groupid.GrpID) (string, error) {
	if !gid.IsValid() {
		return "", errors.New("not a valid group id")
	}

	type groupMbrArg struct {
		msg string
	}

	arg := &groupMbrArg{}

	mdb := db.GetMetaDb()

	mdb.TraversGroupMembers(gid, arg, func(arg, v interface{}) (ret interface{}, err error) {
		gm := v.(*db.GroupMember)
		ag := arg.(*groupMbrArg)

		if ag.msg != "" {
			ag.msg += "\r\n"
		}

		ag.msg += gm.String()

		return nil, nil
	})

	return arg.msg, nil
}
