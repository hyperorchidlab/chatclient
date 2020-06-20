package db

import (
	"github.com/kprc/chat-protocol/groupid"
	"sync"
)

var (
	groupKeyMemLock sync.Mutex
	groupKeyMemDb   map[groupid.GrpID]string
)

func UpdateGroupKeyMem(gid groupid.GrpID, hashKey string) {
	groupKeyMemLock.Lock()
	defer groupKeyMemLock.Unlock()
	if groupKeyMemDb == nil {
		groupKeyMemDb = make(map[groupid.GrpID]string)
	}

	groupKeyMemDb[gid] = hashKey
}

func GetMemGroupKey(gid groupid.GrpID) string {
	if v, ok := groupKeyMemDb[gid]; !ok {
		return ""
	} else {
		return v
	}
}
