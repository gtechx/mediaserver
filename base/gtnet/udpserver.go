package gtnet

import (
	"fmt"
	"net"
)

type GTSendPacket struct {
	buff  []byte
	raddr *net.UDPAddr
}

type GTUdpServer struct {
	OnStart    func()
	OnStop     func()
	OnError    func(int, string)
	OnRecv     func([]byte, *net.UDPAddr)
	OnPreSend  func(*GTSendPacket)
	onPostSend func(*GTSendPacket, int)
	//OnSend func([]byte, *net.UDPAddr)

	conn     *GTUdpConn
	sendChan chan *GTSendPacket
}

func NewUdpServer(ip string, port int) *GTUdpServer {
	return &GTUdpServer{conn: NewUdpConn(ip, port), sendChan: make(chan *GTSendPacket, 1024)}
}

// func NewUdpServer(addr *net.UDPAddr) *UdpServer {
// 	return &UdpServer{IP: addr.IP, Port: addr.Port, addr: addr}
// }

func (g *GTUdpServer) Start() {
	var err error
	err = g.conn.StartListen()

	if err != nil {
		if g.OnError != nil {
			g.OnError(1, "StartListen error:"+err.Error())
		}
		return
	}

	if g.OnStart != nil {
		g.OnStart()
	}
}

func (g *GTUdpServer) Stop() {
	var err error

	err = g.conn.Close()

	if err != nil {
		if g.OnError != nil {
			g.OnError(1, "Stop error:"+err.Error())
		}
	}

	if g.OnStop != nil {
		g.OnStop()
	}
}

func (g *GTUdpServer) Send(packet *GTSendPacket) {
	g.sendChan <- packet
}

func (g *GTUdpServer) startUDPRecv() {
	buffer := make([]byte, 10240)

	for {
		num, raddr, err := g.conn.Recv(buffer)
		if err != nil {
			if g.OnError != nil {
				g.OnError(1, "Recv error:"+err.Error())
			}
			// if g.OnError != nil {
			// 	g.OnError(g.conn, 2, "ReadFromUDP err:"+err.Error())
			// }
			continue
		}

		newbuf := make([]byte, num)
		copy(newbuf, buffer[0:num])
		//newbuf = append(newbuf, buffer[0:num]...)
		g.OnRecv(newbuf, raddr)
	}
}

func (g *GTUdpServer) startUDPSend() {
	for {
		packet := <-g.sendChan

		if g.OnPreSend != nil {
			g.OnPreSend(packet)
		}

		num, err := g.conn.Send(packet.buff, packet.raddr)
		if err != nil {
			fmt.Println("err Send:" + err.Error())
			if g.OnError != nil {
				g.OnError(1, "Send error:"+err.Error())
			}
			return
		}

		if g.onPostSend != nil {
			g.onPostSend(packet, num)
		}
	}
}
