package chatmessage

import (
	"github.com/kprc/chat-protocol/address"
	"github.com/kprc/chatclient/db"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/ssa/interp/testdata/src/fmt"

	"sync"
)

type MsgChannel struct {
	Lock    sync.Mutex
	Running bool
	Msg     chan []*db.FriendMsg
	refresh chan struct{}
	quit    chan struct{}
}

var (
	p2pListenLock sync.Mutex
	p2pListenMem  map[string]*MsgChannel
)

func getChannel(key string) *MsgChannel {
	p2pListenLock.Lock()
	defer p2pListenLock.Unlock()

	if v, ok := p2pListenMem[key]; !ok {
		mc := &MsgChannel{}
		mc.Msg = make(chan []*db.FriendMsg, 100)
		mc.refresh = make(chan struct{}, 100)
		mc.quit = make(chan struct{}, 1)
		p2pListenMem[key] = mc

		return mc
	} else {
		return v
	}
}

func checkRunning(mc *MsgChannel) error {
	mc.Lock.Lock()
	defer mc.Lock.Unlock()
	if mc.Running {
		return errors.New("mc is running")
	}
	mc.Running = true

	return nil
}

func Listen(friend address.ChatAddress) string {
	mc := getChannel(friend.String())

	if mc.Running {
		return "mc is running"
	}

	err := checkRunning(mc)
	if err != nil {
		return err.Error()
	}

	fdb := db.GetMetaDb()
	f, err := fdb.FindFriend(friend)
	if err != nil {
		return "friend not found"
	}

	for {
		select {
		case m := <-mc.Msg:
			for i := 0; i < len(m); i++ {
				msg := m[i]
				s := ""
				if msg.IsOwner {
					s = fmt.Sprintf("%-20s%s", "Mine:", msg.Msg)
				} else {
					s = fmt.Sprintf("%-20s%s", f.AliasName+":", msg.Msg)
				}
				fmt.Println(s)
			}
		case <-mc.refresh:
			//do refresh
		case <-mc.quit:
			mc.Running = false
			return "normal quit"
		}
	}

}
