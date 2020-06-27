package chatmessage

import (
	"errors"
	"fmt"
	"github.com/kprc/chat-protocol/groupid"
	"github.com/kprc/chatclient/app/cmdlistenudp"
	"github.com/kprc/chatclient/db"
	"sync"
	"time"
)

type GMsgChannel struct {
	Lock    sync.Mutex
	Running bool
	Msg     chan []*db.GroupMsg
	refresh chan struct{}
	quit    chan struct{}
	showCnt int
}

var (
	groupListenLock sync.Mutex
	groupListenMem  map[string]*GMsgChannel
)

func init() {
	groupListenMem = make(map[string]*GMsgChannel)
}

func getgChannel(key string) *GMsgChannel {
	groupListenLock.Lock()
	defer groupListenLock.Unlock()

	if v, ok := groupListenMem[key]; !ok {
		mc := &GMsgChannel{}
		mc.Msg = make(chan []*db.GroupMsg, 100)
		mc.refresh = make(chan struct{}, 100)
		mc.quit = make(chan struct{}, 1)
		groupListenMem[key] = mc

		return mc
	} else {
		return v
	}
}

func checkGCRunning(mc *GMsgChannel) error {
	mc.Lock.Lock()
	defer mc.Lock.Unlock()
	if mc.Running {
		return errors.New("mc is running")
	}
	mc.Running = true

	return nil
}

func isGCRunning(mc *GMsgChannel) bool {
	mc.Lock.Lock()
	defer mc.Lock.Unlock()

	return mc.Running

}

func StopGCListen(gid groupid.GrpID) string {
	groupListenLock.Lock()
	defer groupListenLock.Unlock()

	if v, ok := groupListenMem[gid.String()]; !ok {
		return "not found listen thread"
	} else {
		if !isGCRunning(v) {
			return "listen is not running"
		} else {
			v.quit <- struct{}{}
			return "stopping listen group: " + gid.String()
		}

	}
}

func GCListen(gid groupid.GrpID, port string) string {
	mc := getgChannel(gid.String())

	if mc.Running {
		return "mc is running"
	}

	err := checkGCRunning(mc)
	if err != nil {
		return err.Error()
	}

	fdb := db.GetMetaDb()
	_, err = fdb.FindGroup(gid)
	if err != nil {
		return "group not found"
	}

	c := cmdlistenudp.NewCmdUdpClient(port)
	if err = c.Start(); err != nil {
		return "Can't start tunnel"
	}

	tc := time.NewTicker(time.Second)

	for {
		select {
		case m := <-mc.Msg:
			for i := 0; i < len(m); i++ {
				msg := m[i]

				s := fmt.Sprintf("%-20s%s", msg.Speak+":", DecryptGMsg(gid, msg.AesKey, msg.Msg))

				c.Write([]byte(s))
			}
		case <-tc.C:
			mc.refreshFriend(gid)
		case <-mc.refresh:
			mc.refreshFriend(gid)
		case <-mc.quit:
			c.Write([]byte("====quit===="))
			time.Sleep(time.Second)
			tc.Stop()
			mc.Running = false
			c.Close()
			return "normal quit"
		}
	}

}

func DecryptGMsg(gid groupid.GrpID, aesk string, cipherMsg string) (plainMsg string) {
	//cfg := config.GetCCC()

	//ciphertxt := base58.Decode(cipherMsg)
	//
	//aesk, err := chatcrypt.GenerateAesKey(friend.ToPubKey(), cfg.PrivKey)
	//if err != nil {
	//	return ""
	//}
	var plaintxt []byte
	//plaintxt, err = chatcrypt.Decrypt(aesk, []byte(ciphertxt))

	return string(plaintxt)
}

func (mc *GMsgChannel) refreshFriend(gid groupid.GrpID) {

	var (
		begin int
	)

	gmdb := db.GetGroupMsgDb()
	v, err := gmdb.FindLatest(gid)
	if err != nil {
		begin = 0
	} else {
		begin = v.Counter + 1
	}

	err = FetchGroupMsg(gid, begin, 20)
	if err != nil {
		fmt.Println(err)
		//return
	}

	fms, err := gmdb.FindMsg(gid, mc.showCnt, 20)
	if err != nil {
		fmt.Println(err)
		return
	}

	//for i := 0; i < len(fms); i++ {
	//	log.Println(fms[i].String())
	//}

	if len(fms) <= 0 {
		return
	}

	mc.Msg <- fms

	mc.showCnt = fms[len(fms)-1].LCnt + 1

}
