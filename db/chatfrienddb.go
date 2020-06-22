package db

import (
	"encoding/json"
	"github.com/kprc/chatclient/config"
	"github.com/kprc/nbsnetwork/db"
	"github.com/kprc/nbsnetwork/hdb"
	"sync"
)

type FriendHistoryDB struct {
	hdb.HistoryDBIntf
	lock   sync.Mutex
	cursor *db.DBCusor
}

type FriendMsg struct {
	IsOwner bool   `json:"is_owner"`
	Msg     string `json:"msg"`
	Counter int    `json:"cnt"`
	LCnt    int    `json:"l_cnt"`
}

func newFriendMsgDb() *FriendHistoryDB {
	db := hdb.New(2000, config.GetCCC().GetChatFriendPath()).Load()

	return &FriendHistoryDB{HistoryDBIntf: db}
}

var (
	friendMsgDB     *FriendHistoryDB
	friendMsgDBLock sync.Mutex
)

func GetFriendMsgDb() *FriendHistoryDB {
	if friendMsgDB == nil {
		friendMsgDBLock.Lock()
		defer friendMsgDBLock.Unlock()

		if friendMsgDB == nil {
			friendMsgDB = newFriendMsgDb()
		}

	}

	return friendMsgDB
}

func (fhdb *FriendHistoryDB) Insert(peerPk string, isOwner bool, msg string, cnt int) error {
	fhdb.lock.Lock()
	defer fhdb.lock.Unlock()

	m := &FriendMsg{IsOwner: isOwner, Msg: msg, Counter: cnt}

	d, _ := json.Marshal(&m)

	fhdb.HistoryDBIntf.Insert(peerPk, string(d))

	return nil
}

func (fhdb *FriendHistoryDB) FindMsg(peerPk string, begin, n int) ([]*FriendMsg, error) {
	fhdb.lock.Lock()
	defer fhdb.lock.Unlock()

	v, err := fhdb.HistoryDBIntf.Find(peerPk, begin, n)
	if err != nil {
		return nil, err
	}

	var r []*FriendMsg

	for i := 0; i < len(v); i++ {
		vv := v[i]

		fm := &FriendMsg{}

		json.Unmarshal([]byte(vv.V), fm)

		fm.LCnt = vv.Cnt

		r = append(r, fm)
	}

	return r, nil

}
