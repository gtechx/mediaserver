package gtnet

import (
	"fmt"
	"net"
	"strconv"
)

type UdpServer struct {
	OnStart func(*net.UDPConn)
	OnClose func(*net.UDPConn)
	OnRecv  func(*net.UDPConn) ([]byte, int)
	OnSend  func(*net.UDPConn) []byte

	IP   string
	Port int
	addr *net.UDPAddr
	conn *net.UDPConn
}

func NewUdpServer(ip string, port int) *UdpServer {
	return &UdpServer{IP: ip, Port: port}
}

// func NewUdpServer(addr *net.UDPAddr) *UdpServer {
// 	return &UdpServer{IP: addr.IP, Port: addr.Port, addr: addr}
// }

func (u *UdpServer) Start() {
	var err error
	u.addr, _ = net.ResolveUDPAddr("udp", u.IP+":"+strconv.Itoa(u.Port))
	u.conn, err = net.ListenUDP("udp", u.addr)
	if err != nil {
		fmt.Println(err)
		return
	}

	if u.OnStart != nil {
		u.OnStart(u.conn)
	}
}

func (u *UdpServer) Close() {
	var err error

	if u.OnClose != nil {
		u.OnClose(u.conn)
	}

	err = u.conn.Close()

	if err != nil {
		fmt.Println(err)
	}
}
