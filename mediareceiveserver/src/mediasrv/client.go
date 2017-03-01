package mediasrv

import (
	"fmt"
	"net"
)

type client struct {
	Ip   string       `json:"ip"`
	Port int          `json:"port"`
	conn *net.UDPConn `json:"-"`
}

func newClient(ip string, port int, conn *net.UDPConn) *client {
	c := client{ip, port, conn}
	fmt.Println("new client ", ip)
	return &c
}
