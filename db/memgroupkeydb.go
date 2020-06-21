package db

import (
	"errors"
	"github.com/kprc/chat-protocol/address"
	"github.com/kprc/chat-protocol/groupid"
	"github.com/kprc/chatclient/chatcrypt"
	"github.com/kprc/chatclient/config"
	"sync"
)

type GroupKeyMemInfo struct {
	HashKey string
	AesKey  []byte
}

var (
	groupKeyMemLock sync.Mutex
	groupKeyMemDb   map[groupid.GrpID]*GroupKeyMemInfo
)

func UpdateGroupKeyMem(gid groupid.GrpID, hashKey string) {
	groupKeyMemLock.Lock()
	defer groupKeyMemLock.Unlock()
	//if groupKeyMemDb == nil {
	//	groupKeyMemDb = make(map[groupid.GrpID]*GroupKeyMemInfo)
	//}

	m := &GroupKeyMemInfo{HashKey: hashKey}

	groupKeyMemDb[gid] = m
}

func GetMemGroupKey(gid groupid.GrpID) *GroupKeyMemInfo {
	groupKeyMemLock.Lock()
	defer groupKeyMemLock.Unlock()

	if v, ok := groupKeyMemDb[gid]; !ok {
		return nil
	} else {
		return v
	}
}

func GetMemGroupAesKey(gid groupid.GrpID) (aesk []byte, err error) {
	groupKeyMemLock.Lock()
	defer groupKeyMemLock.Unlock()

	if v, ok := groupKeyMemDb[gid]; !ok {
		return nil, errors.New("Not found")
	} else {
		gkdb := GetChatGrpKeysDb()
		gks := gkdb.Find(v.HashKey)

		aesk, err = chatcrypt.DeriveGroupKey2(config.GetCCC().PrivKey, gks.GroupKeys, gks.PubKeys)
		if err != nil {
			return nil, err
		}
		v.AesKey = aesk
		return aesk, nil
	}
}

type FriendKeyMemInfo struct {
	FriendAddr address.ChatAddress
	aesKey     []byte
}

var (
	friendKeyMemLock sync.Mutex
	friendKeyMemDb   map[address.ChatAddress][]byte
)

func GetMemFriendAesKey(fid address.ChatAddress) (aesk []byte, err error) {
	if !fid.IsValid() {
		return nil, errors.New("Not a correct chat address")
	}
	friendKeyMemLock.Lock()
	defer friendKeyMemLock.Unlock()

	if v, ok := friendKeyMemDb[fid]; !ok {
		k, err := chatcrypt.GenerateAesKey(fid.GetBytes(), config.GetCCC().PrivKey)
		if err != nil {
			return nil, err
		}
		friendKeyMemDb[fid] = k
		return k, nil

	} else {
		return v, nil
	}

}

func init() {
	if groupKeyMemDb == nil {
		groupKeyMemDb = make(map[groupid.GrpID]*GroupKeyMemInfo)
	}

	if friendKeyMemDb == nil {
		friendKeyMemDb = make(map[address.ChatAddress][]byte)
	}

}
