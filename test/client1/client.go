package main

import (
	//"encoding/json"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
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
type roomInfo struct {
	Id        string    `json:"id"`
	Ip        string    `json:"ip"`
	Port      int       `json:"port"`
	Clients   []*client `json:"subroom"`
	ErrorCode int       `json:"errorcode"`
	Error     string    `json:"error"`
}

var loginedroom roomInfo

func startUDPCon() {
	fmt.Println("logining..." + "http://" + sip + ":12345/login?useraccount=1001")
	resp, err := http.Get("http://" + sip + ":12345/login?useraccount=1001")
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
		resp, err = http.Get("http://" + sip + ":12345/create?sessionid=" + strconv.FormatInt(info.SessionId, 10))
		defer resp.Body.Close()
		if err != nil {
			// handle error
			fmt.Println(err.Error())
			//io.WriteString(rw, "{\"error\":\"http error\"}")
			return
		}
		body, err = ioutil.ReadAll(resp.Body)

		loginedroom.ErrorCode = -1
		json.Unmarshal(body, &loginedroom)

		if loginedroom.ErrorCode == -1 {
			//conn, err := net.Dial("udp", "127.0.0.1:4040")
			fmt.Println("create room success")
			udpAddr, err := net.ResolveUDPAddr("udp", loginedroom.Ip+":"+strconv.Itoa(loginedroom.Port))

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
		}
	}
}

func processUDPWrite(conn *net.UDPConn) {
	//var content string
	f, err := os.Open("male.wav")
	if err != nil {
		fmt.Println("open file err:" + err.Error())
		return
	}

	finfo, err := f.Stat()

	if err != nil {
		fmt.Println("Stat file err:" + err.Error())
		return
	}
	fmt.Println("file size:", finfo.Size())
	pretime := time.Now().UnixNano() / 1000 / 1000
	f.Seek(0, 44)
	for {
		//fmt.Scanln(&content)
		nowtime := time.Now().UnixNano() / 1000 / 1000
		if nowtime-pretime > 100 {
			pretime = nowtime
			wavbuf := make([]byte, 1600)
			num, err := f.Read(wavbuf)

			if err == io.EOF {
				f.Seek(0, 44)
				//break
			} else if err != nil {
				fmt.Println("read file err:" + err.Error())
				return
			}
			var datasize int32
			//var uid int64
			var dtype byte

			//uid = 1012
			dtype = 2
			datasize = int32(num)

			bytesBuffer := bytes.NewBuffer([]byte{})
			binary.Write(bytesBuffer, binary.LittleEndian, datasize)
			sendbuf := bytesBuffer.Bytes()

			bytesBuffer = bytes.NewBuffer([]byte{})
			binary.Write(bytesBuffer, binary.LittleEndian, info.SessionId)
			sendbuf = append(sendbuf, bytesBuffer.Bytes()...)

			bytesBuffer = bytes.NewBuffer([]byte{})
			binary.Write(bytesBuffer, binary.LittleEndian, dtype)
			sendbuf = append(sendbuf, bytesBuffer.Bytes()...)

			sendbuf = append(sendbuf, wavbuf[0:num]...)

			conn.Write(sendbuf)
			fmt.Println("send msg is ", wavbuf[0:num])
		}
	}
	f.Close()
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
			fmt.Println("connect rs server ok, start sending data...")
			go processUDPWrite(conn)
		} else {
			fmt.Println("connect rs server failed")
		}
	}
}
