package cmdlistenudp

import (
	"net"
	"strconv"
)

type CmdUdpClient struct {
	addr net.UDPAddr
	conn *net.UDPConn
}

func NewCmdUdpClient(port string) *CmdUdpClient {
	cuc:=&CmdUdpClient{}
	iport,_:=strconv.Atoi(port)
	cuc.addr = net.UDPAddr{IP: net.ParseIP("127.0.0.1"),Port: iport}

	return cuc
}

func (cuc *CmdUdpClient)Start() error  {
	laddr:=net.UDPAddr{}
	var err error
	cuc.conn,err = net.DialUDP("udp",&laddr,&cuc.addr)
	if err!=nil{
		return err
	}

	return nil
}

func (cuc *CmdUdpClient)Write(buf []byte) error {
	_,err:=cuc.conn.Write(buf)

	return err
}

func (cuc *CmdUdpClient)Close()  {
	cuc.conn.Close()
}