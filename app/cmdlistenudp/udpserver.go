package cmdlistenudp

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"
)

type CmdUpdServer struct {
	addr net.UDPAddr
	conn *net.UDPConn
	wg   sync.WaitGroup
}

func (cus *CmdUpdServer) Serve() error {

	var err error
	cus.conn, err = net.ListenUDP("udp", &cus.addr)
	if err != nil {
		return err
	}
	defer cus.conn.Close()

	cus.wg.Add(1)
	go func(cus *CmdUpdServer) {
		defer cus.wg.Done()
		buf := make([]byte, 20480)
		for {
			n, _, err := cus.conn.ReadFromUDP(buf)
			if err != nil {
				return
			}

			if string(buf[:n]) == "====quit====" {
				return
			}

			fmt.Println(string(buf[:n]))
		}

	}(cus)

	cus.wg.Wait()

	return nil
}

func NewUdpServer(port int) *CmdUpdServer {
	cus := &CmdUpdServer{}

	cus.addr = net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: port}

	return cus
}

func RandPort() int {
	rand.Seed(time.Now().UnixNano())

	n := rand.Intn(1000)

	return 51000 + n
}
