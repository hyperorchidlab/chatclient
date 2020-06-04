package config

import (
	"encoding/json"
	"github.com/kprc/chat-protocol/protocol"
	"github.com/kprc/nbsnetwork/tools"
	"log"
)

type KeyJson struct {
	PubKey    string `json:"pub_key"`
	CipherKey string `json:"cipher_key"`
}

type UserIdentify struct {
	SP protocol.SignPack `json:"sp"`
	KJ KeyJson           `json:"kj"`
}

func IsUserIdentifyReceived() bool {
	cfg := GetCCC()

	ufpath := cfg.GetUserFile()

	if tools.FileExists(ufpath) {
		return true
	}

	return false
}

func SaveUserIdentify(sp *protocol.SignPack) {
	cfg := GetCCC()

	if sp == nil || cfg.PrivKey == nil {
		panic("No Private In Memory")
	}

	ui := &UserIdentify{}

	ui.SP = *sp

	data, err := tools.OpenAndReadAll(cfg.GetKeyPath())
	if err != nil {
		log.Fatal("Load From key file error")
		return
	}

	kj := &KeyJson{}

	err = json.Unmarshal(data, kj)
	if err != nil {
		log.Fatal("Load From json error")
		return
	}

	ui.KJ = *kj

	data, err = json.Marshal(*ui)
	if err != nil {
		log.Fatal("Save json error")
		return
	}
	err = tools.Save2File(data, cfg.GetUserFile())
	if err != nil {
		log.Fatal("Save to file error")
		return
	}

	cfg.SP = sp

}

func loadUserIdentify() {

	cfg := GetCCC()
	data, err := tools.OpenAndReadAll(cfg.GetUserFile())
	if err != nil {
		log.Fatal("Load userfile ")
		return
	}

	ui := &UserIdentify{}

	err = json.Unmarshal(data, ui)
	if err != nil {
		log.Fatal("unmarshal user identity data failed")
		return
	}

	cfg.SP = &ui.SP

	return
}
