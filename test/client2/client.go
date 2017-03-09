package main

import (
	//"encoding/json"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	//"net/http"
)

// type client struct {
// 	Ip   string `json:"ip"`
// 	Port int    `json:"port"`
// }

// type roomInfo struct {
// 	Id      string             `json:"id"`
// 	Ip      string             `json:"ip"`
// 	Port    int                `json:"port"`
// 	Subroom map[string]*client `json:"subroom"`
// }

var c chan int
var sip string
var sport int

func main() {
	c := make(chan int)

	pip := flag.String("ip", "192.168.96.124", "ip address")
	pport := flag.Int("port", 20001, "port")
	flag.Parse()
	sip = *pip
	sport = *pport
	fmt.Println("ip:", sip)
	fmt.Println("port:", sport)

	go startUDPCon()
	_ = <-c
}

type loginInfo struct {
	SessionId int64  `json:"uid"`
	ErrorCode int    `json:"errorcode"`
	Error     string `json:"error"`
}

var info loginInfo

type client struct {
	Ip   string `json:"ip"`
	Port int    `json:"port"`
}
type room struct {
	Id        string    `json:"id"`
	Ip        string    `json:"ip"`
	Port      int       `json:"port"`
	Clients   []*client `json:"subroom"`
	ErrorCode int       `json:"errorcode"`
	Error     string    `json:"error"`
}

var loginedroom room

type roomInfo struct {
	RoomType    string `json:"type"` //主播，自由
	HasPassword int    `json:"haspassword"`
	password    string `json:"-"`
	IsPublic    int    `json:"ispublic"`
	PRoom       *room  `json:"room"`
	sessionId   string `json:"-"`
}

type errorInfo struct {
	ErrorCode int    `json:"errorcode"`
	Error     string `json:"error"`
}

var listroominfo []roomInfo
var errInfo errorInfo

func startUDPCon() {
	fmt.Println("logining..." + "http://" + sip + ":12345/login?useraccount=1002")
	resp, err := http.Get("http://" + sip + ":12345/login?useraccount=1002")
	defer resp.Body.Close()
	if err != nil {
		// handle error
		fmt.Println(err.Error())
		//io.WriteString(rw, "{\"error\":\"http error\"}")
		return
	}
	body, err := ioutil.ReadAll(resp.Body)

	fmt.Println(string(body))
	info.ErrorCode = -1
	// var rinfo roomInfo
	json.Unmarshal(body, &info)

	// b, _ := json.Marshal(rinfo)
	fmt.Println(info)

	if info.ErrorCode == -1 {
		//create room
		fmt.Println("creating room...")
		resp, err = http.Get("http://" + sip + ":12345/listrooms?sessionid=" + strconv.FormatInt(info.SessionId, 10))
		defer resp.Body.Close()
		if err != nil {
			// handle error
			fmt.Println(err.Error())
			//io.WriteString(rw, "{\"error\":\"http error\"}")
			return
		}
		body, err = ioutil.ReadAll(resp.Body)

		//listroominfo.ErrorCode = -1
		errInfo.ErrorCode = -1
		json.Unmarshal(body, &errInfo)

		if errInfo.ErrorCode == -1 {
			json.Unmarshal(body, &listroominfo)
			fmt.Println(string(body))
			fmt.Println(listroominfo)

			//conn, err := net.Dial("udp", "127.0.0.1:4040")
			fmt.Println(listroominfo)
			udpAddr, err := net.ResolveUDPAddr("udp", listroominfo[0].PRoom.Clients[0].Ip+":"+strconv.Itoa(listroominfo[0].PRoom.Clients[0].Port))

			//udp连接
			conn, err := net.DialUDP("udp", nil, udpAddr)
			//defer conn.Close()
			if err != nil {
				fmt.Println(err)
				return
			}

			buf := make([]byte, 4)

			bytesBuffer := bytes.NewBuffer([]byte{})
			binary.Write(bytesBuffer, binary.LittleEndian, info.SessionId)
			buf = append(buf, bytesBuffer.Bytes()...)

			buf = append(buf, []byte{0}...)
			fmt.Println(buf)
			conn.Write(buf)
			//conn.Write([]byte("Hello world!"))

			//go processUDPRead(conn)
			go processUDPRead(conn)
		} else {
			fmt.Println("error:", errInfo)
		}
	}
}

func processUDPWrite(conn *net.UDPConn) {
	var content string
	for {
		fmt.Scanln(&content)
		datasize := len(content)
		var uid int64
		uid = 1010
		var dtype byte
		dtype = 2

		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, binary.LittleEndian, datasize)
		sendbuf := bytesBuffer.Bytes()

		bytesBuffer = bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, binary.LittleEndian, uid)
		sendbuf = append(sendbuf, bytesBuffer.Bytes()...)

		bytesBuffer = bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, binary.LittleEndian, dtype)
		sendbuf = append(sendbuf, bytesBuffer.Bytes()...)

		sendbuf = append(sendbuf, []byte(content)...)

		conn.Write(sendbuf)
		fmt.Println("send msg is " + content)
	}

}

func processUDPRead(conn *net.UDPConn) {
	for {
		allbuf := make([]byte, 2048)

		var datasize int32
		var uid int64

		_, err := conn.Read(allbuf[0:])
		if err != nil {
			fmt.Println("err:" + err.Error())
			continue
		}

		dsizebuf := allbuf[0:4]
		uidbuf := allbuf[4:12]
		btype := allbuf[12:13]

		b_buf := bytes.NewBuffer(dsizebuf)
		binary.Read(b_buf, binary.LittleEndian, &datasize)
		fmt.Println("data size is ", datasize)

		b_buf = bytes.NewBuffer(uidbuf)
		binary.Read(b_buf, binary.LittleEndian, &uid)
		fmt.Println("uid is ", uid)

		fmt.Println("type is ", btype[0])

		if btype[0] == 200 {
			fmt.Println("login to bs server ok")
		}

		databuf := allbuf[13 : 13+datasize]

		fmt.Println("recv msg is ", databuf)
	}
}
