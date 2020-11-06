package chatmeta

import (
	"encoding/json"
	"errors"
	"github.com/btcsuite/btcutil/base58"
	"github.com/hyperorchidlab/chat-protocol/address"
	"github.com/hyperorchidlab/chat-protocol/groupid"
	"github.com/hyperorchidlab/chat-protocol/protocol"
	"github.com/hyperorchidlab/chatclient/chatcrypt"
	"github.com/hyperorchidlab/chatclient/config"
	"github.com/hyperorchidlab/chatclient/db"
	"github.com/hyperorchidlab/chatserver/app/cmdcommon"
	"log"
	"strconv"
)

func CreateGroup(groupName string) error {

	gd := &protocol.GroupDesc{}
	gd.GroupAlias = groupName
	gd.GroupID = groupid.NewGroupId().String()
	gd.SendTime = cmdcommon.GetNowMsTime()

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

	resp, stat, err = cmdcommon.Post1(config.GetCCC().GetCmdUrl(), string(d2s), false)
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
		return errors.New("create group failed, cipher text is none")
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
		mdb.AddGroupMember(resp.GCI.GID, cfg.SP.SignText.AliasName, address.ChatAddress(cfg.SP.SignText.CPubKey), 0, resp.GCI.CreateTime)

		return nil
	}

	return errors.New("create group failed")

}

func DelGroup(gid groupid.GrpID) error {

	gd := &protocol.GroupDesc{}
	gd.GroupID = gid.String()
	gd.SendTime = cmdcommon.GetNowMsTime()

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

	resp, stat, err = cmdcommon.Post1(config.GetCCC().GetCmdUrl(), string(d2s), false)
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

func groupIsOwner(gid groupid.GrpID) bool {
	mdb := db.GetMetaDb()

	g, err := mdb.FindGroup(gid)
	if err != nil {
		log.Println("check group is owner failed")
		return false
	}

	return g.IsOwner
}

func getAllAddress(gid groupid.GrpID, friendPK string, drop bool) ([]address.ChatAddress, error) {

	mdb := db.GetMetaDb()

	type chatAddrs struct {
		addrs []address.ChatAddress
	}

	arg := &chatAddrs{}

	mdb.TraversGroupMembers(gid, arg, func(arg, v interface{}) (ret interface{}, err error) {
		m := v.(*db.GroupMember)
		a := arg.(*chatAddrs)

		a.addrs = append(a.addrs, m.Addr)

		return nil, nil
	})

	idx := 0

	for i := 0; i < len(arg.addrs); i++ {
		if arg.addrs[i].String() == friendPK {
			if !drop {
				return nil, errors.New("Address is exists")
			} else {
				idx = i
				break
			}
		}
	}

	if !drop {
		arg.addrs = append(arg.addrs, address.ChatAddress(friendPK))
	} else {

		al := len(arg.addrs)
		if al == idx {
			return nil, errors.New("not found address")
		}

		arg.addrs[idx] = arg.addrs[al-1]
		arg.addrs = arg.addrs[:al-1]
	}

	return arg.addrs, nil
}

func address2PKs(addrs []address.ChatAddress) (pkeys []string) {
	for i := 0; i < len(addrs); i++ {
		pkeys = append(pkeys, addrs[i].TrimPrefix())
	}
	return
}

func address2PKBytes(addrs []address.ChatAddress) (pkBytes [][]byte) {
	for i := 0; i < len(addrs); i++ {
		pkBytes = append(pkBytes, addrs[i].GetBytes())
	}
	return
}

func bytesArrays2StringArrays(byteArrs [][]byte) (stringArrs []string) {
	for i := 0; i < len(byteArrs); i++ {
		stringArrs = append(stringArrs, base58.Encode(byteArrs[i]))
	}

	return
}

func genGroupKeys(gid groupid.GrpID, friendPK string, drop bool) (pkeys []string, gkeys []string, err error) {
	addrs, err := getAllAddress(gid, friendPK, drop)
	if err != nil {
		return
	}
	pkBytes := address2PKBytes(addrs)

	cfg := config.GetCCC()

	var (
		gkBytes [][]byte
		aesk    []byte
	)

	aesk, gkBytes, err = chatcrypt.GenGroupAesKey(cfg.PrivKey, pkBytes)
	if err != nil {
		return
	}

	gkdb := db.GetChatGrpKeysDb()
	memkey := gkdb.Insert(gkBytes, pkBytes)
	gmemdb := db.GetGrouKeyMemDb()
	gmemdb.UpdateGroupKeyMem(gid, memkey, aesk)

	pkeys = bytesArrays2StringArrays(pkBytes)
	gkeys = bytesArrays2StringArrays(gkBytes)

	return

}

func JoinGroup(gid groupid.GrpID, friendPk string) error {

	if !gid.IsValid() {
		return errors.New("group id is not corrected")
	}
	if !address.ChatAddress(friendPk).IsValid() {
		return errors.New("user id not correct")
	}

	if !groupIsOwner(gid) {
		return errors.New("group is not owner")
	}

	var (
		err error
		pks []string
		gks []string
	)
	pks, gks, err = genGroupKeys(gid, friendPk, false)
	if err != nil {
		errors.New("generate group key failed")
	}

	cfg := config.GetCCC()

	uc := &protocol.UserCommand{}
	uc.Op = protocol.JoinGroup
	uc.SP = *cfg.SP

	gmd := &protocol.GroupMemberDesc{}
	gmd.GroupID = gid.String()
	gmd.Friend = friendPk
	gmd.SendTime = cmdcommon.GetNowMsTime()
	gmd.Pubkeys = pks
	gmd.GKeys = gks

	serverPub := address.ChatAddress(cfg.SP.SignText.SPubKey).ToPubKey()
	aesk, _ := chatcrypt.GenerateAesKey(serverPub, cfg.PrivKey)

	data, _ := json.Marshal(*gmd)

	ciphertxt, _ := chatcrypt.Encrypt(aesk, data)

	uc.CipherTxt = base58.Encode(ciphertxt)

	d2s, _ := json.Marshal(uc)

	var (
		resp string
		stat int
	)
	log.Println(string(d2s))

	resp, stat, err = cmdcommon.Post1(config.GetCCC().GetCmdUrl(), string(d2s), false)
	if err != nil {
		return err
	}
	if stat != 200 {
		return errors.New("Get Error Stat Code:" + strconv.Itoa(stat))
	}

	log.Println(resp)

	reply := &protocol.UCReply{}
	json.Unmarshal([]byte(resp), reply)

	if reply.CipherTxt == "" {
		return errors.New("join group failed, cipher text is empty")
	}

	if reply.ResultCode == 0 && reply.OP == protocol.JoinGroup {
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

		gkdb := db.GetChatGrpKeysDb()
		hks := gkdb.Insert2(gmd.GKeys, gmd.Pubkeys)
		if hks != gi.GKeyHash {
			log.Println(" group key hash not corrected")
		}
		gkmdb := db.GetGrouKeyMemDb()
		gkmdb.UpdateGroupKeyMem(gi.GID, hks, nil)

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

	if !groupIsOwner(gid) {
		return errors.New("group is not owner")
	}
	var (
		err error
		pks []string
		gks []string
	)
	pks, gks, err = genGroupKeys(gid, friendPk, false)
	if err != nil {
		errors.New("generate group key failed")
	}

	cfg := config.GetCCC()

	uc := &protocol.UserCommand{}
	uc.Op = protocol.QuitGroup
	uc.SP = *cfg.SP

	gmd := &protocol.GroupMemberDesc{}
	gmd.GroupID = gid.String()
	gmd.Friend = friendPk
	gmd.SendTime = cmdcommon.GetNowMsTime()
	gmd.Pubkeys = pks
	gmd.GKeys = gks

	serverPub := address.ChatAddress(cfg.SP.SignText.SPubKey).ToPubKey()
	aesk, _ := chatcrypt.GenerateAesKey(serverPub, cfg.PrivKey)

	data, _ := json.Marshal(*gmd)

	ciphertxt, _ := chatcrypt.Encrypt(aesk, data)

	uc.CipherTxt = base58.Encode(ciphertxt)

	d2s, _ := json.Marshal(uc)

	var (
		resp string
		stat int
	)
	log.Println(string(d2s))

	resp, stat, err = cmdcommon.Post1(config.GetCCC().GetCmdUrl(), string(d2s), false)
	if err != nil {
		return err
	}
	if stat != 200 {
		return errors.New("Get Error Stat Code:" + strconv.Itoa(stat))
	}

	log.Println(resp)

	reply := &protocol.UCReply{}
	json.Unmarshal([]byte(resp), reply)

	if reply.CipherTxt == "" {
		return errors.New("quit group failed, cipher text is not equal")
	}

	if reply.ResultCode == 0 && reply.OP == protocol.QuitGroup {

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
		mdb.DelGroupMember(gid, address.ChatAddress(friendPk))

		gkdb := db.GetChatGrpKeysDb()
		hashk := gkdb.Insert2(gmd.GKeys, gmd.Pubkeys)

		if hashk != resp.GMAI.GKeyHash {
			log.Println("group key is not equals")
		}

		gkmdb := db.GetGrouKeyMemDb()
		gkmdb.UpdateGroupKeyMem(resp.GMAI.GID, hashk, nil)

		//db.UpdateGroupKeyMem(resp.GMAI.GID, hashk)

		return nil
	}

	return errors.New("quit group failed")
}
