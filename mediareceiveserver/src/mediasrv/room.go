package mediasrv

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
)

type Room struct {
	Id       string                 `json:"id"`
	Ip       string                 `json:"ip"`
	Port     int                    `json:"port"`
	conn     *net.UDPConn           `json:"-"`
	iclients map[int64]*net.UDPAddr `json:"-"`
	Clients  map[string]*client     `json:"subroom"`
}

type clientroom struct {
	Id   string
	Ip   string
	Port int
}

func NewRoom(id string, ip string, port int, clientdata string) *Room {
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

	return &Room{id, ip, port, nil, make(map[int64]*net.UDPAddr), cmap}
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
			//input client
			fmt.Println("input client connected:" + raddr.String())
			r.iclients[uid] = raddr
		} else {
			//trans to client servers
			// databuf := make([]byte, datasize)
			// _, raddr, _ := conn.ReadFromUDP(databuf[0:])
			fmt.Println("input client data:" + raddr.String())
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
