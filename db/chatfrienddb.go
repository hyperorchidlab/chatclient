package db

import (
	"encoding/json"
	"fmt"
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
	LCnt    int    `json:"-"`
}

func (fm *FriendMsg) String() string {
	s := ""
	s += fmt.Sprintf("IsOwner:%-8v", fm.IsOwner)
	s += fmt.Sprintf("Msg: %s-32", fm.Msg)
	s += fmt.Sprintf("Cnt: %-8d", fm.Counter)
	s += fmt.Sprintf("LC: %-8d", fm.LCnt)
	return s
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

func (fhdb *FriendHistoryDB) FindLatest(key string) (*FriendMsg, error) {
	fhdb.lock.Lock()
	defer fhdb.lock.Unlock()

	if v, err := fhdb.HistoryDBIntf.FindLatest(key); err != nil {
		return nil, err
	} else {
		fm := &FriendMsg{}
		json.Unmarshal([]byte(v.V), fm)
		fm.LCnt = v.Cnt

		return fm, nil
	}

}
