package gtnet

import (
	"fmt"
	"net"
	"strconv"
)

type GTUdpConn struct {
	//OnConnected func(*net.UDPConn)
	//OnClose     func()
	//OnError     func(*net.UDPConn, int, string)
	//OnRecv  func([]byte, int, *net.UDPConn, *net.UDPAddr)
	//OnSend  func([]byte, int, *net.UDPConn, *net.UDPAddr)

	IP   string
	Port int
	addr *net.UDPAddr
	conn *net.UDPConn
}

func NewUdpConn(ip string, port int) *GTUdpConn {
	return &GTUdpConn{IP: ip, Port: port}
}

func (g *GTUdpConn) StartListen() error {
	var err error

	g.addr, err = net.ResolveUDPAddr("udp", g.IP+":"+strconv.Itoa(g.Port))

	if err != nil {
		fmt.Println("GTUdpConn StartListen ResolveUDPAddr error:" + err.Error())
		// if g.OnError != nil {
		// 	g.OnError(g.conn, 1, "ResolveUDPAddr error:"+err.Error())
		// }
		return err
	}

	g.conn, err = net.ListenUDP("udp", g.addr)

	if err != nil {
		fmt.Println("GTUdpConn StartListen ListenUDP error:" + err.Error())
		// if g.OnError != nil {
		// 	g.OnError(g.conn, 1, "ListenUDP error:"+err.Error())
		// }
		return err
	}

	return nil
}

func (g *GTUdpConn) Connect() error {
	var err error

	g.addr, err = net.ResolveUDPAddr("udp", g.IP+":"+strconv.Itoa(g.Port))

	if err != nil {
		fmt.Println("GTUdpConn Connect ResolveUDPAddr error:" + err.Error())
		return err
	}

	g.conn, err = net.DialUDP("udp", nil, g.addr)
	if err != nil {
		fmt.Println("GTUdpConn Connect DialUDP error:" + err.Error())
		return err
	}

	return nil
}

func (g *GTUdpConn) Close() error {
	var err error

	err = g.conn.Close()

	if err != nil {
		fmt.Println("GTUdpConn Close error:" + err.Error())
		return err
	}

	return nil
}

func (g *GTUdpConn) Send(buff []byte, raddr *net.UDPAddr) (int, error) {
	return g.conn.WriteToUDP(buff, raddr)
}

func (g *GTUdpConn) Recv(buff []byte) (int, *net.UDPAddr, error) {
	return g.conn.ReadFromUDP(buff)
}
