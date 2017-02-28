package mediasrv

import (
	"fmt"
	"net"
)

type client struct {
	ip   string
	port int
	conn *net.UDPConn
}

func newClient(ip string) *client {
	c := client{ip: ip}
	fmt.Println("new client ", ip)
	return &c
}
