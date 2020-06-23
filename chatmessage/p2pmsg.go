package chatmessage

import (
	"encoding/json"
	"errors"
	"github.com/btcsuite/btcutil/base58"
	"github.com/kprc/chat-protocol/address"
	"github.com/kprc/chat-protocol/protocol"
	"github.com/kprc/chatclient/chatcrypt"
	"github.com/kprc/chatclient/config"
	"github.com/kprc/chatclient/db"
	"github.com/rickeyliao/ServiceAgent/common"
	"log"
	"strconv"
)

func SendP2pMsg(friend address.ChatAddress, message string) error {
	cfg := config.GetCCC()

	uc := protocol.UserCommand{}
	uc.Op = protocol.StoreP2pMsg
	uc.SP = *cfg.SP

	pm := &protocol.P2pMsg{}
	pm.PeerPk = friend.String()
	pm.MyPk = address.ToAddress(cfg.PubKey).String()

	ciphermsg, err := db.EncryptP2pMsg(message, friend)
	if err != nil {
		return err
	}

	pm.Msg = ciphermsg

	serverPub := address.ChatAddress(cfg.SP.SignText.SPubKey).ToPubKey()
	aesk, _ := chatcrypt.GenerateAesKey(serverPub, cfg.PrivKey)

	data, _ := json.Marshal(*pm)

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
		return errors.New("send p2p msg failed, cipher text is none")
	}

	if reply.ResultCode == 0 && reply.OP == protocol.StoreP2pMsg {
		return nil
	}

	return errors.New("Send message failed")
}

func FetchP2pMessage(friend address.ChatAddress, begin, n int) error {
	cfg := config.GetCCC()

	uc := protocol.UserCommand{}
	uc.Op = protocol.FetchP2pMsg
	uc.SP = *cfg.SP

	pf := &protocol.P2pMsgFetch{PeerPk: friend.String(), Begin: begin, Count: n}

	serverPub := address.ChatAddress(cfg.SP.SignText.SPubKey).ToPubKey()
	aesk, _ := chatcrypt.GenerateAesKey(serverPub, cfg.PrivKey)

	data, _ := json.Marshal(*pf)

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
		return errors.New("fetch p2p msg failed, cipher text is none")
	}

	if reply.ResultCode == 0 && reply.OP == protocol.FetchP2pMsg {
		cipherBytes := base58.Decode(reply.CipherTxt)
		var plaintxt []byte
		plaintxt, err = chatcrypt.Decrypt(aesk, cipherBytes)
		resp := &protocol.P2pMsgFetchResp{}
		err = json.Unmarshal(plaintxt, &resp.Msg)
		if err != nil {
			log.Println("fetch friend message info error")
			return nil
		}
		for i := 0; i < len(resp.Msg); i++ {
			lm := resp.Msg[i]

			fmdb := db.GetFriendMsgDb()

			var isOwner bool

			if lm.PubKey == cfg.SP.SignText.CPubKey {
				isOwner = true
			}

			fmdb.Insert(friend.String(), isOwner, lm.Msg, lm.Cnt)

		}

		return nil
	}

	return errors.New("Fetch p2p message failed")

}
