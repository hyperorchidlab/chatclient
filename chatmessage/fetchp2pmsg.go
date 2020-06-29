package chatmessage

import (
	"fmt"
	"github.com/kprc/chat-protocol/address"
	"github.com/kprc/chatclient/app/cmdlistenudp"
	"github.com/kprc/chatclient/db"
	"github.com/pkg/errors"
	"time"

	"sync"
)

type MsgChannel struct {
	Lock    sync.Mutex
	Running bool
	Msg     chan []*db.FriendMsg
	refresh chan struct{}
	quit    chan struct{}
	showCnt int
}

var (
	p2pListenLock sync.Mutex
	p2pListenMem  map[string]*MsgChannel
)

func init() {
	p2pListenMem = make(map[string]*MsgChannel)
}

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

func isRunning(mc *MsgChannel) bool {
	mc.Lock.Lock()
	defer mc.Lock.Unlock()

	return mc.Running

}

func StopListen(friend address.ChatAddress) string {
	p2pListenLock.Lock()
	defer p2pListenLock.Unlock()

	if v, ok := p2pListenMem[friend.String()]; !ok {
		return "not found listen thread"
	} else {
		if !isRunning(v) {
			return "listen is not running"
		} else {
			v.quit <- struct{}{}
			return "stopping listen friend: " + friend.String()
		}

	}
}

func Listen(friend address.ChatAddress, port string) string {
	mc := getChannel(friend.String())

	err := checkRunning(mc)
	if err != nil {
		return err.Error()
	}

	fdb := db.GetMetaDb()
	f, err := fdb.FindFriend(friend)
	if err != nil {
		return "friend not found"
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
				s := ""

				var plainTxt string
				plainTxt, err = db.DecryptP2pMsg(friend, msg.Msg)
				if err != nil {
					continue
				}

				if msg.IsOwner {
					s = fmt.Sprintf("%-20s%s", "Mine:", plainTxt)
				} else {
					s = fmt.Sprintf("%-20s%s", f.AliasName+":", plainTxt)
				}
				c.Write([]byte(s))
			}
		case <-tc.C:
			mc.refreshFriend(friend)
		case <-mc.refresh:
			mc.refreshFriend(friend)
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

//func DecryptMsg(friend address.ChatAddress, cipherMsg string) (plainMsg string) {
//	cfg := config.GetCCC()
//
//	ciphertxt := base58.Decode(cipherMsg)
//
//	aesk, err := chatcrypt.GenerateAesKey(friend.ToPubKey(), cfg.PrivKey)
//	if err != nil {
//		return ""
//	}
//	var plaintxt []byte
//	plaintxt, err = chatcrypt.Decrypt(aesk, []byte(ciphertxt))
//
//	return string(plaintxt)
//}

func (mc *MsgChannel) refreshFriend(friend address.ChatAddress) {

	var (
		begin int
	)

	fmdb := db.GetFriendMsgDb()
	v, err := fmdb.FindLatest(friend.String())
	if err != nil {
		begin = 0
	} else {
		begin = v.Counter + 1
	}

	err = FetchP2pMessage(friend, begin, 20)
	if err != nil {
		fmt.Println(err)
		//return
	}

	fms, err := fmdb.FindMsg(friend.String(), mc.showCnt, 20)
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
