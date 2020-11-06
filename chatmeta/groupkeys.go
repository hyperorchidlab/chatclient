package chatmeta

import (
	"encoding/json"
	"errors"
	"github.com/btcsuite/btcutil/base58"
	"github.com/hyperorchidlab/chat-protocol/address"
	"github.com/hyperorchidlab/chat-protocol/protocol"
	"github.com/hyperorchidlab/chatclient/chatcrypt"
	"github.com/hyperorchidlab/chatclient/config"
	"github.com/hyperorchidlab/chatclient/db"
	"github.com/hyperorchidlab/chatserver/app/cmdcommon"
	"log"
	"strconv"
)

func FetchGroupKey(hahsKey string) error {
	cfg := config.GetCCC()

	uc := &protocol.UserCommand{}
	uc.Op = protocol.FetchGrpKeys
	uc.SP = *cfg.SP

	req := &protocol.GroupKeyFetchReq{}

	req.GKI.IndexKey = hahsKey

	serverPub := address.ChatAddress(cfg.SP.SignText.SPubKey).ToPubKey()
	aesk, _ := chatcrypt.GenerateAesKey(serverPub, cfg.PrivKey)

	data, _ := json.Marshal(req.GKI)

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
		return errors.New("fetch group encryption key failed, cipher text is none")
	}

	if reply.ResultCode == 0 && reply.OP == protocol.FetchGrpKeys {
		cipherBytes := base58.Decode(reply.CipherTxt)
		var plaintxt []byte
		plaintxt, err = chatcrypt.Decrypt(aesk, cipherBytes)
		resp := &protocol.GroupKeyFetchResp{}
		err = json.Unmarshal(plaintxt, &resp.GKs)
		if err != nil {
			log.Println("group create info error")
			return nil
		}

		gkdb := db.GetChatGrpKeysDb()
		gkdb.Insert2(resp.GKs.GroupKeys, resp.GKs.PubKeys)

		return nil
	}

	return errors.New("fetch group encryption key failed")

}
func FetchGroupKey2(hahsKey string) (*protocol.GroupKeys, error) {
	cfg := config.GetCCC()

	uc := &protocol.UserCommand{}
	uc.Op = protocol.FetchGrpKeys
	uc.SP = *cfg.SP

	req := &protocol.GroupKeyFetchReq{}

	req.GKI.IndexKey = hahsKey

	serverPub := address.ChatAddress(cfg.SP.SignText.SPubKey).ToPubKey()
	aesk, _ := chatcrypt.GenerateAesKey(serverPub, cfg.PrivKey)

	data, _ := json.Marshal(req.GKI)

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
		return nil, err
	}
	if stat != 200 {
		return nil, errors.New("Get Error Stat Code:" + strconv.Itoa(stat))
	}

	log.Println(resp)

	reply := &protocol.UCReply{}
	err = json.Unmarshal([]byte(resp), reply)
	if err != nil {
		return nil, err
	}
	if reply.CipherTxt == "" {
		return nil, errors.New("fetch group encryption key failed, cipher text is none")
	}

	if reply.ResultCode == 0 && reply.OP == protocol.FetchGrpKeys {
		cipherBytes := base58.Decode(reply.CipherTxt)
		var plaintxt []byte
		plaintxt, err = chatcrypt.Decrypt(aesk, cipherBytes)
		resp := &protocol.GroupKeyFetchResp{}
		err = json.Unmarshal(plaintxt, &resp.GKs)
		if err != nil {
			log.Println("group create info error")
			return nil, err
		}

		gkdb := db.GetChatGrpKeysDb()
		gkdb.Insert2(resp.GKs.GroupKeys, resp.GKs.PubKeys)

		return &protocol.GroupKeys{GroupKeys: resp.GKs.GroupKeys, PubKeys: resp.GKs.PubKeys}, nil

	}

	return nil, errors.New("fetch group encryption key failed")
}
