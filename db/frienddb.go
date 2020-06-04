package db

import (
	"github.com/kprc/chat-protocol/groupid"
	"github.com/kprc/nbsnetwork/db"
	"sync"
)

type ClientFriendDb struct {
	db.NbsDbInter
	dbLock sync.Mutex
	cursor *db.DBCusor
}

var (
	cfStore     *ClientFriendDb
	cfStoreLock sync.Mutex
)

type Friend struct {
	AliasName string `json:"alias_name"`
	PubKey    string `json:"pub_key"`
	AddTime   int64  `json:"add_time"`
	Agree     bool   `json:"agree"`
}

type Group struct {
	GroupName string        `json:"group_name"`
	GroupId   groupid.GrpID `json:"group_id"`
	JoinTime  int64         `json:"join_time"`
	Owner     string        `json:"owner"`
}
