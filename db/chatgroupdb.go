package db

import (
	"encoding/json"
	"github.com/kprc/chat-protocol/address"
	"github.com/kprc/chat-protocol/groupid"
	"github.com/kprc/chatclient/config"
	"github.com/kprc/nbsnetwork/db"
	"github.com/kprc/nbsnetwork/hdb"
	"sync"
)

type GroupHistoryDB struct {
	hdb.HistoryDBIntf
	lock   sync.Mutex
	cursor *db.DBCusor
}

type GroupMsg struct {
	Speak   string `json:"speak"`
	Msg     string `json:"msg"`
	AesKey  string `json:"aes_key"`
	Counter int    `json:"cnt"`
	LCnt    int    `json:"-"`
}

func newGroupMsgDb() *GroupHistoryDB {

	db := hdb.New(2000, config.GetCCC().GetChatGroupPath()).Load()

	return &GroupHistoryDB{HistoryDBIntf: db}
}

var (
	groupMsgDB     *GroupHistoryDB
	groupMsgDBLock sync.Mutex
)

func GetGroupMsgDb() *GroupHistoryDB {
	if groupMsgDB == nil {
		groupMsgDBLock.Lock()
		defer groupMsgDBLock.Unlock()

		if groupMsgDB == nil {
			groupMsgDB = newGroupMsgDb()
		}

	}

	return groupMsgDB
}

func (ghdb *GroupHistoryDB) Insert(gid groupid.GrpID, speak address.ChatAddress, aesk, msg string, cnt int) error {
	ghdb.lock.Lock()
	defer ghdb.lock.Unlock()

	m := &GroupMsg{Speak: speak.String(), Msg: msg, Counter: cnt, AesKey: aesk}

	d, _ := json.Marshal(&m)

	ghdb.HistoryDBIntf.Insert(gid.String(), string(d))

	return nil
}

func (ghdb *GroupHistoryDB) FindMsg(gid groupid.GrpID, begin, n int) ([]*GroupMsg, error) {
	ghdb.lock.Lock()
	defer ghdb.lock.Unlock()

	v, err := ghdb.HistoryDBIntf.Find(gid.String(), begin, n)
	if err != nil {
		return nil, err
	}

	var r []*GroupMsg

	for i := 0; i < len(v); i++ {
		vv := v[i]

		gm := &GroupMsg{}

		json.Unmarshal([]byte(vv.V), gm)

		gm.LCnt = vv.Cnt

		r = append(r, gm)
	}

	return r, nil

}

func (ghdb *GroupHistoryDB) FindLatest(gid groupid.GrpID) (*GroupMsg, error) {
	ghdb.lock.Lock()
	defer ghdb.lock.Unlock()

	if v, err := ghdb.HistoryDBIntf.FindLatest(gid.String()); err != nil {
		return nil, err
	} else {
		gm := &GroupMsg{}
		json.Unmarshal([]byte(v.V), gm)
		gm.LCnt = v.Cnt

		return gm, nil
	}
}
