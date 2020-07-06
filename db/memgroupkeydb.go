package db

import (
	"errors"
	"github.com/btcsuite/btcutil/base58"
	"github.com/kprc/chat-protocol/address"
	"github.com/kprc/chat-protocol/groupid"
	"github.com/kprc/chat-protocol/protocol"
	"github.com/kprc/chatclient/chatcrypt"
	"github.com/kprc/chatclient/config"
	"github.com/kprc/chatclient/msgdrive"
	"sync"
)


type GroupKeyPair struct {
	HK     string
	AesKey []byte
}

type GroupKeyMap struct {
	KeyMap map[string][]byte
	CurKP  *GroupKeyPair
}

type GroupAesKeyMem struct {
	KM map[groupid.GrpID]*GroupKeyMap
}

var (
	groupKeyMemLock sync.Mutex
	groupKeyMemDb   *GroupAesKeyMem
)

func GetGrouKeyMemDb() *GroupAesKeyMem {

	if groupKeyMemDb != nil {
		return groupKeyMemDb
	}

	groupKeyMemLock.Lock()
	defer groupKeyMemLock.Unlock()
	if groupKeyMemDb != nil {
		return groupKeyMemDb
	}

	groupKeyMemDb = &GroupAesKeyMem{KM: make(map[groupid.GrpID]*GroupKeyMap)}

	return groupKeyMemDb
}

func (gakm *GroupAesKeyMem) UpdateGroupKeyMem(gid groupid.GrpID, hashKey string, aesk []byte) {
	groupKeyMemLock.Lock()
	defer groupKeyMemLock.Unlock()

	gi, ok := groupKeyMemDb.KM[gid]
	if !ok {
		gi = &GroupKeyMap{}
		gi.KeyMap = make(map[string][]byte)
		groupKeyMemDb.KM[gid] = gi
	}

	gi.KeyMap[hashKey] = aesk

	if gi.CurKP == nil {
		gi.CurKP = &GroupKeyPair{}
	}
	gi.CurKP.AesKey = aesk
	gi.CurKP.HK = hashKey

}

func (gakm *GroupAesKeyMem) GetMemGroupHashKey(gid groupid.GrpID) (string, error) {
	groupKeyMemLock.Lock()
	defer groupKeyMemLock.Unlock()

	if v, ok := groupKeyMemDb.KM[gid]; !ok {
		return "", errors.New("no groupkey")
	} else {
		if v.CurKP == nil {
			return "", errors.New("no groupkey")
		}
		return v.CurKP.HK, nil
	}

}

func (gakm *GroupAesKeyMem) GetMemGroupAesKey(gid groupid.GrpID) (aesk []byte, err error) {
	groupKeyMemLock.Lock()
	defer groupKeyMemLock.Unlock()

	var (
		v  *GroupKeyMap
		ok bool
	)

	v, ok = groupKeyMemDb.KM[gid]
	if !ok {
		return nil, errors.New("Not found")
	}
	if v.CurKP == nil || v.CurKP.HK == "" {
		return nil, errors.New("No hot hash key")
	}

	if v.CurKP.AesKey != nil {
		return v.CurKP.AesKey, nil
	}

	gkdb := GetChatGrpKeysDb()
	gks := gkdb.Find(v.CurKP.HK)

	if gks == nil {
		var protcolgks *protocol.GroupKeys
		protcolgks, err = msgdrive.DriveGroupKey(v.CurKP.HK)
		if err != nil {
			return nil, err
		}

		gks.GroupKeys = protcolgks.GroupKeys
		gks.PubKeys = protcolgks.PubKeys

	}

	aesk, err = getAesKey(gid, gks)

	v.KeyMap[v.CurKP.HK] = aesk
	v.CurKP.AesKey = aesk

	return aesk, nil
}

func getAesKey(gid groupid.GrpID, gks *GroupKeys) ([]byte, error) {

	var (
		aesk []byte
		err  error
	)

	cfg := config.GetCCC()

	mdb := GetMetaDb()
	if mdb.IsOwnerGroup(gid) {
		aesk, _, err = chatcrypt.GenGroupAesKey2(cfg.PrivKey, gks.PubKeys)
		if err != nil {
			return nil, err
		}
	} else {

		aesk, err = chatcrypt.DeriveGroupKey2(cfg.PrivKey, gks.GroupKeys, gks.PubKeys)
		if err != nil {
			return nil, err
		}

	}

	return aesk, err
}

func EncryptGroupMsg(message string, gid groupid.GrpID) (string, error) {
	gkdb := GetGrouKeyMemDb()
	aesk, err := gkdb.GetMemGroupAesKey(gid)
	if err != nil {
		return "", err
	}

	var cipherTxt []byte
	cipherTxt, err = chatcrypt.Encrypt(aesk, []byte(message))
	if err != nil {
		return "", err
	}

	return base58.Encode(cipherTxt), nil

}

func DecryptGroupMsg(cipherTxt string, hashk string, gid groupid.GrpID) (string, error) {

	var aesk []byte

	v, ok := groupKeyMemDb.KM[gid]
	if !ok {
		v = &GroupKeyMap{KeyMap: make(map[string][]byte)}
		groupKeyMemDb.KM[gid] = v
	}

	if v.CurKP != nil {
		if v.CurKP.HK == hashk {
			aesk = v.CurKP.AesKey
		}
	} else {
		if v.KeyMap != nil {
			aesk = v.KeyMap[hashk]
		}
	}
	var err error
	if aesk == nil || len(aesk) == 0 {

		grpkdb := GetChatGrpKeysDb()
		gks := grpkdb.Find(hashk)
		if gks == nil {
			var ks *protocol.GroupKeys
			ks, err = msgdrive.DriveGroupKey(hashk)
			if err != nil {
				return "", err
			}
			gks = &GroupKeys{PubKeys: ks.PubKeys, GroupKeys: ks.GroupKeys}
		}

		aesk, err = getAesKey(gid, gks)
		if err != nil {
			return "", err
		}
		if v.KeyMap == nil {
			v.KeyMap = make(map[string][]byte)
		}
		v.KeyMap[hashk] = aesk
		if v.CurKP == nil {
			v.CurKP = &GroupKeyPair{}
		}
		v.CurKP.HK = hashk
		v.CurKP.AesKey = aesk
	}

	cipherBytes := base58.Decode(cipherTxt)

	var plainTxt []byte
	plainTxt, err = chatcrypt.Decrypt(aesk, cipherBytes)
	if err != nil {
		return "", err
	}
	return string(plainTxt), nil
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

func EncryptP2pMsg(message string, friend address.ChatAddress) (string, error) {
	aesk, err := GetMemFriendAesKey(friend)
	if err != nil {
		return "", err
	}
	var cipherTxt []byte
	cipherTxt, err = chatcrypt.Encrypt(aesk, []byte(message))
	if err != nil {
		return "", err
	}

	return base58.Encode(cipherTxt), nil
}

func DecryptP2pMsg(friend address.ChatAddress, cipherTxt string) (string, error) {
	aesk, err := GetMemFriendAesKey(friend)
	if err != nil {
		return "", err
	}

	cipherBytes := base58.Decode(cipherTxt)
	var plainTxt []byte

	plainTxt, err = chatcrypt.Decrypt(aesk, cipherBytes)
	if err != nil {
		return "", err
	}

	return string(plainTxt), nil
}

func init() {

	if friendKeyMemDb == nil {
		friendKeyMemDb = make(map[address.ChatAddress][]byte)
	}

}
