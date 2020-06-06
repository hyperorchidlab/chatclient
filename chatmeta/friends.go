package chatmeta

import (
	"github.com/kprc/chat-protocol/address"
	"github.com/pkg/errors"
	"github.com/kprc/chatclient/config"
	"github.com/kprc/chat-protocol/protocol"
	"github.com/kprc/chatclient/chatcrypt"
	"encoding/json"
	"github.com/kprc/nbsnetwork/tools"
	"github.com/btcsuite/btcutil/base58"
	"log"
	"github.com/rickeyliao/ServiceAgent/common"
	"strconv"
)

func AddFriend(addr address.ChatAddress) error  {
	if !addr.IsValid(){
		return errors.New("addr is error")
	}

	cfg:=config.GetCCC()

	localaddr:=address.ToAddress(cfg.PubKey)

	if localaddr == addr{
		return errors.New("don't add self as friend")
	}

	fd:=&protocol.FriendDesc{}
	fd.PeerPubKey = addr.String()
	fd.SendTime = tools.GetNowMsTime()

	cmd:=protocol.UserCommand{}
	cmd.Op = protocol.AddFriend
	cmd.SP = *cfg.SP

	serverPub:=address.ChatAddress(cfg.SP.SignText.SPubKey).ToPubKey()
	aesk,_:=chatcrypt.GenerateAesKey(serverPub,cfg.PrivKey)

	data,_:=json.Marshal(*fd)

	ciphertxt,_ := chatcrypt.Encrypt(aesk,data)

	cmd.CipherTxt = base58.Encode(ciphertxt)

	d2s,_:=json.Marshal(cmd)

	var (
		resp string
		stat int
		err error
	)
	log.Println(string(d2s))

	resp, stat, err = common.Post1(config.GetCCC().GetCmdUrl(), string(d2s), false)
	if err != nil {
		return err
	}
	if stat != 200 {
		return errors.New("Get Error Stat Code:"+strconv.Itoa(stat))
	}

	log.Println(resp)

	reply:=&protocol.UCReply{}
	json.Unmarshal([]byte(resp),reply)

	if reply.CipherTxt != cmd.CipherTxt{
		return errors.New("add friend failed, cipher text is not equal")
	}

	if reply.ResultCode == 0 || reply.OP == protocol.AddFriend{
		return nil
	}

	return errors.New("add friend failed")

}