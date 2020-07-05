package chatmessage

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
	"github.com/rickeyliao/ServiceAgent/common"
	"log"
	"strconv"
)

func SendGroupMsg(gid groupid.GrpID, msg string) error {
	cfg := config.GetCCC()
	uc := protocol.UserCommand{}
	uc.Op = protocol.StoreGMsg
	uc.SP = *cfg.SP

	gm := &protocol.GroupMsg{}
	//gm.Speek = address.ChatAddress(uc.SP.SignText.CPubKey)
	gm.Gid = gid
	gm.Msg = msg

	var err error

	gmkdb := db.GetGrouKeyMemDb()

	gm.AesHash, err = gmkdb.GetMemGroupHashKey(gid)
	if err != nil {
		return err
	}

	gm.Msg, err = db.EncryptGroupMsg(msg, gid)
	if err != nil {
		return err
	}

	serverPub := address.ChatAddress(cfg.SP.SignText.SPubKey).ToPubKey()
	aesk, _ := chatcrypt.GenerateAesKey(serverPub, cfg.PrivKey)

	data, _ := json.Marshal(*gm)

	ciphertxt, _ := chatcrypt.Encrypt(aesk, data)

	uc.CipherTxt = base58.Encode(ciphertxt)

	d2s, _ := json.Marshal(uc)

	var (
		resp string
		stat int
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
		return errors.New("send group msg failed, cipher text is none")
	}

	if reply.ResultCode == 0 && reply.OP == protocol.StoreGMsg {
		return nil
	}

	return errors.New("Send group message failed")
}

func FetchGroupMsg(gid groupid.GrpID, begin, n int) error {
	cfg := config.GetCCC()
	uc := protocol.UserCommand{}
	uc.Op = protocol.FetchGMsg
	uc.SP = *cfg.SP

	gf := &protocol.GMsgFetch{Gid: gid, Begin: begin, Count: n}

	serverPub := address.ChatAddress(cfg.SP.SignText.SPubKey).ToPubKey()
	aesk, _ := chatcrypt.GenerateAesKey(serverPub, cfg.PrivKey)

	data, _ := json.Marshal(*gf)

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
	if reply.CipherTxt == "" {
		return errors.New("fetch group msg failed, cipher text is none")
	}

	if reply.ResultCode == 0 && reply.OP == protocol.FetchGMsg {
		cipherBytes := base58.Decode(reply.CipherTxt)
		var plaintxt []byte
		plaintxt, err = chatcrypt.Decrypt(aesk, cipherBytes)
		resp := &protocol.GMsgFetchResp{}
		err = json.Unmarshal(plaintxt, &resp.GMsg)
		if err != nil {
			log.Println("fetch friend message info error")
			return nil
		}
		for i := 0; i < len(resp.GMsg.LM); i++ {
			lm := resp.GMsg.LM[i]

			gmdb := db.GetGroupMsgDb()
			gmdb.Insert(resp.GMsg.Gid, lm.Speek, lm.AesHash, lm.Msg, lm.Cnt,lm.UCnt)

		}

		return nil
	}

	return errors.New("Fetch group message failed")
}
