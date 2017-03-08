package main

import (
	//"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"time"
)

type Room struct {
	Id        string                  `json:"id"`
	Ip        string                  `json:"ip"`
	Port      int                     `json:"port"`
	conn      *net.UDPConn            `json:"-"`
	parent    *net.UDPAddr            `json:"-"`
	clients   map[string]*net.UDPAddr `json:"-"`
	loginMaps map[string]*net.UDPAddr `json:"-"`
	scip      string                  `json:"-"`
	scport    int                     `json:"-"`
}

func NewRoom(id string, ip string, portid int, scip string, scport int) *Room {
	fmt.Println("new room ", id)
	return &Room{id, ip, portid, nil, nil, make(map[string]*net.UDPAddr), make(map[string]*net.UDPAddr), scip, scport}
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
		allbuf := make([]byte, 2048)

		var datasize int32
		var uid string

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

		uid = string(uidbuf)
		// b_buf = bytes.NewBuffer(uidbuf)
		// binary.Read(b_buf, binary.LittleEndian, &uid)
		// fmt.Println("uid is ", uid)

		if btype[0] == 0 {
			//user client
			fmt.Println("user client logining:" + raddr.String())
			r.loginMaps[uid] = raddr
			go r.doCheckLogin(uid, raddr)
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
			fmt.Println(time.Now().Format("2006-01-02 15:04:05") + "parent data:" + raddr.String())
			sendbuf := make([]byte, 0)
			sendbuf = append(sendbuf, allbuf[0:13+datasize]...)
			go r.doUDPWrite(sendbuf, uid)
		}
	}
}

func (r *Room) doUDPWrite(buf []byte, uid string) {
	for id, udpaddr := range r.clients {
		if id == uid {
			continue
		}
		_, err := r.conn.WriteToUDP(buf, udpaddr)
		if err != nil {
			fmt.Println("err doUDPWrite:" + err.Error())
		}
	}
}

type loginCBInfo struct {
	Ok        string
	ErrorCode int
	Error     string
}

func (r *Room) doCheckLogin(uid string, raddr *net.UDPAddr) {
	resp, err := http.Get("http://" + r.scip + ":" + strconv.Itoa(r.scport) + "/checklogin?srvtype=bs&id=" + r.Id + "&sessionid=" + uid)

	if err != nil {
		// handle error
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	var info loginCBInfo
	info.ErrorCode = -1
	json.Unmarshal(body, &info)

	var datasize int32
	var dtype byte

	datasize = int32(0)

	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.LittleEndian, datasize)
	sendbuf := bytesBuffer.Bytes()

	sendbuf = append(sendbuf, []byte(uid)...)

	if info.ErrorCode == -1 {
		fmt.Println("user client logined:" + raddr.String())

		if _, ok := r.loginMaps[uid]; ok {
			r.clients[uid] = r.loginMaps[uid]
		}
		dtype = 200

	} else {
		fmt.Println("user client login failed:" + raddr.String())
		dtype = 201
	}

	bytesBuffer = bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.LittleEndian, dtype)
	sendbuf = append(sendbuf, bytesBuffer.Bytes()...)

	_, err = r.conn.WriteToUDP(sendbuf, raddr)
	if err != nil {
		fmt.Println("err doCheckLogin:" + err.Error())
	}

	delete(r.loginMaps, uid)

	fmt.Println(string(body))
}
