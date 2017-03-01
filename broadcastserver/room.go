package main

import (
	//"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
)

type Room struct {
	Id      string                 `json:"id"`
	Ip      string                 `json:"ip"`
	Port    int                    `json:"port"`
	conn    *net.UDPConn           `json:"-"`
	parent  *net.UDPAddr           `json:"-"`
	clients map[int64]*net.UDPAddr `json:"-"`
}

func NewRoom(id string, ip string, portid int) *Room {
	fmt.Println("new room ", id)
	return &Room{id, ip, portid, nil, nil, make(map[int64]*net.UDPAddr)}
}

func (r *Room) Start() {
	go r.startUDPServer()
}

func (r *Room) startUDPServer() {
	fmt.Println("starting udp server for room on port:" + strconv.Itoa(r.Port))
	udpaddr, _ := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(r.Port))
	conn, err := net.ListenUDP("udp", udpaddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("room udp server start ok")
	r.conn = conn

	fmt.Println(udpaddr.String())

	go r.startUDPRead()
}

func (r *Room) startUDPRead() {
	conn := r.conn
	//reader := bufio.NewReader(conn)
	for {
		allbuf := make([]byte, 128)

		var datasize int32
		var uid int64

		_, raddr, err := conn.ReadFromUDP(allbuf[0:])
		if err != nil {
			fmt.Println("err:" + err.Error())
			continue
		}

		buf := allbuf[0:4]
		uidbuf := allbuf[4:12]
		btype := allbuf[12:13]

		b_buf := bytes.NewBuffer(buf)
		binary.Read(b_buf, binary.LittleEndian, &datasize)
		fmt.Println("data size is ", datasize)

		b_buf = bytes.NewBuffer(uidbuf)
		binary.Read(b_buf, binary.LittleEndian, &uid)
		fmt.Println("uid is ", uid)

		if btype[0] == 0 {
			//user client
			fmt.Println("user client connected:" + raddr.String())
			r.clients[uid] = raddr
		} else if btype[0] == 1 {
			//parent udp server connect
			fmt.Println("parent connect:" + raddr.String())
			r.parent = raddr
		} else {
			//parent udp servers data
			// databuf := make([]byte, datasize)
			// _, raddr, _ := conn.ReadFromUDP(databuf[0:])
			// allbuf := make([]byte, 0)
			// allbuf = append(allbuf, buf...)
			// allbuf = append(allbuf, uidbuf...)
			// allbuf = append(allbuf, btype...)
			// allbuf = append(allbuf, databuf...)
			fmt.Println("parent data:" + raddr.String())
			go r.doUDPWrite(allbuf)
		}
	}
}

func (r *Room) doUDPWrite(buf []byte) {
	for _, udpaddr := range r.clients {
		_, err := r.conn.WriteToUDP(buf, udpaddr)
		if err != nil {
			fmt.Println("err doUDPWrite:" + err.Error())
		}
	}
}
