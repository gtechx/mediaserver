package mediasrv

import (
	"fmt"
	"net"
	"strconv"
)

type Room struct {
	id      string
	port    int
	conn    *net.UDPConn
	clients map[string]*client
}

func NewRoom(id string, port int) *Room {
	fmt.Println("new room ", id)
	return &Room{id, port, nil, make(map[string]*client)}
}

func (r *Room) Start() {
	udpaddr, _ := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(r.port))
	conn, err := net.ListenUDP("udp", udpaddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	r.conn = conn

	fmt.Println(udpaddr.String())

	for {
		var buf [20]byte

		n, raddr, err := conn.ReadFromUDP(buf[0:])
		if err != nil {
			continue
		}

		fmt.Println("msg is ", string(buf[0:n]))

		//WriteToUDP
		//func (c *UDPConn) WriteToUDP(b []byte, addr *UDPAddr) (int, error)
		_, err = conn.WriteToUDP([]byte("nice to see u:"+string(buf[0:n])), raddr)
		if err != nil {
			fmt.Println("err writetoudp:" + err.Error())
		}
	}
}
