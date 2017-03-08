package mediasrv

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
	iclients  map[string]*net.UDPAddr `json:"-"`
	Clients   map[string]*client      `json:"subroom"`
	loginMaps map[string]*net.UDPAddr `json:"-"`
	scip      string                  `json:"-"`
	scport    int                     `json:"-"`
}

type clientroom struct {
	Id   string
	Ip   string
	Port int
}

func NewRoom(id string, ip string, port int, clientdata string, scip string, scport int) *Room {
	fmt.Println("new room ", id)
	fmt.Println(clientdata)
	cmap := make(map[string]*client)
	var croom clientroom
	json.Unmarshal([]byte(clientdata), &croom)

	fmt.Println("connecting bs server:" + croom.Ip + ":" + strconv.Itoa(croom.Port))
	udpAddr, _ := net.ResolveUDPAddr("udp", croom.Ip+":"+strconv.Itoa(croom.Port))
	//udp连接
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		fmt.Println("connect bs server failed:" + err.Error())
		return nil
	}
	fmt.Println("connect bs server ok")
	buf := make([]byte, 13)
	buf[12] = 1
	conn.Write(buf)

	cmap[croom.Id] = newClient(croom.Ip, croom.Port, conn)

	return &Room{id, ip, port, nil, make(map[string]*net.UDPAddr), cmap, make(map[string]*net.UDPAddr), scip, scport}
}

func (r *Room) Start() {
	go r.startUDPServer()
}

func (r *Room) startUDPServer() {
	udpaddr, _ := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(r.Port))
	conn, err := net.ListenUDP("udp", udpaddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	r.conn = conn

	fmt.Println(udpaddr.String())

	go r.startUDPRead()
}

func (r *Room) startUDPRead() {
	conn := r.conn
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
			//input client
			//fmt.Println("input client connected:" + raddr.String())
			//r.iclients[uid] = raddr
			fmt.Println("user client logining:" + raddr.String())
			r.loginMaps[uid] = raddr
			go r.doCheckLogin(uid, raddr)
		} else {
			//trans to client servers
			// databuf := make([]byte, datasize)
			// _, raddr, _ := conn.ReadFromUDP(databuf[0:])
			fmt.Println(time.Now().Format("2006-01-02 15:04:05") + "input client data:" + raddr.String())
			// allbuf := make([]byte, 0)
			// allbuf = append(allbuf, buf...)
			// allbuf = append(allbuf, uidbuf...)
			// allbuf = append(allbuf, btype...)
			// allbuf = append(allbuf, databuf...)
			// allbuf := make([]byte, len(buf)+len(uidbuf)+len(btype)+datasize)
			// copy(allbuf,)
			// append(allbuf, buf, uidbuf, databuf,...)
			sendbuf := make([]byte, 0)
			sendbuf = append(sendbuf, allbuf[0:13+datasize]...)
			go r.doUDPWrite(sendbuf)
		}
	}
}

func (r *Room) doUDPWrite(buf []byte) {
	for _, value := range r.Clients {
		_, err := value.conn.Write(buf)
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
			r.iclients[uid] = r.loginMaps[uid]
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
