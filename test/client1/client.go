package main

import (
	//"encoding/json"
	"../../common"
	"../../common/error"
	"../../common/protocol"
	"../../common/room"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"
	"utils"
	//"unsafe"
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

	pip := flag.String("ip", "192.168.1.50", "ip address")
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
	SessionId uint64 `json:"sessionid"`
}

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
	gterr := new(gterror.Error)
	json.Unmarshal(body, gterr)

	if gterr.ErrorCode != 0 {
		return
	}

	info := new(loginInfo)
	json.Unmarshal(body, info)

	// b, _ := json.Marshal(rinfo)
	fmt.Println(info)

	//create room
	fmt.Println("creating room...")
	fmt.Println("http://" + sip + ":12345/create?sessionid=" + utils.Uint64ToStr(info.SessionId) + "&roomname=1001")
	resp, err = http.Get("http://" + sip + ":12345/create?sessionid=" + utils.Uint64ToStr(info.SessionId) + "&roomname=1001")
	defer resp.Body.Close()
	if err != nil {
		// handle error
		fmt.Println(err.Error())
		//io.WriteString(rw, "{\"error\":\"http error\"}")
		return
	}
	body, err = ioutil.ReadAll(resp.Body)

	fmt.Println(string(body))
	json.Unmarshal(body, gterr)
	if gterr.ErrorCode != 0 {
		return
	}

	room := new(gtroom.SCRoom)
	json.Unmarshal(body, room)

	//conn, err := net.Dial("udp", "127.0.0.1:4040")
	fmt.Println("create room success")
	udpAddr, err := net.ResolveUDPAddr("udp", room.RSList[0])

	//udp连接
	conn, err := net.DialUDP("udp", nil, udpAddr)
	//defer conn.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	//login
	proto := new(gtprotocol.ReqLoginProtocol)
	proto.DataSize = 8
	proto.MsgType = common.MSG_REQ_LOGIN
	copy(proto.RoomName[:], []byte("1001"))
	//strbyte := make([]byte, 32)
	//copy(strbyte, proto.RoomName[:])
	index := bytes.IndexByte(proto.RoomName[:], 0)
	//rbyf_pn := proto.RoomName[0:index]
	//str := *(*string)(unsafe.Pointer(&strbyte))
	//str = strings.TrimSpace(str)
	//str = strings.Replace(str, " ", "", -1)
	fmt.Println(string(proto.RoomName[0:index]) + "55")
	proto.SessionId = info.SessionId
	buff := proto.ToBytes()
	fmt.Println("buff:", buff)

	conn.Write(buff)
	//conn.Write([]byte("Hello world!"))

	//go processUDPRead(conn)
	go processUDPRead(conn)
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
			proto := new(gtprotocol.DataTransProtocol)
			proto.DataSize = int32(num)
			proto.MsgType = common.MSG_DATA_TRANS
			copy(proto.RoomName[:], []byte("1001"))
			proto.Data = wavbuf

			buff := proto.ToBytes()
			fmt.Println("send msg is ", buff)
			conn.Write(buff)
		}
	}
	f.Close()
}

func processUDPRead(conn *net.UDPConn) {
	for {
		allbuf := make([]byte, 2048)

		_, err := conn.Read(allbuf[0:])
		if err != nil {
			fmt.Println("err:" + err.Error())
			continue
		}

		proto := new(gtprotocol.RetLoginProtocol)
		proto.Parse(allbuf)

		if proto.Result == 1 {
			fmt.Println("login rs server ok, start sending data...")
			go processUDPWrite(conn)
		} else {
			fmt.Println("login rs server failed")
		}
	}
}
