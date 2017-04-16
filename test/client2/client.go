package main

import (
	//"encoding/json"
	"../../common"
	"../../common/error"
	"../../common/protocol"
	"../../common/room"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"utils"
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

var roomMap map[string]*gtroom.SCRoom

func main() {
	c := make(chan int)
	roomMap = make(map[string]*gtroom.SCRoom)

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
	gterr := new(gterror.Error)
	json.Unmarshal(body, gterr)

	if gterr.ErrorCode != 0 {
		return
	}

	info := new(loginInfo)
	json.Unmarshal(body, info)

	fmt.Println(info)

	//list room
	fmt.Println("creating room...")
	resp, err = http.Get("http://" + sip + ":12345/listrooms?sessionid=" + utils.Uint64ToStr(info.SessionId))
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

	err = json.Unmarshal(body, &roomMap)

	if err != nil {
		fmt.Println(err.Error())
	}

	var conn *net.UDPConn
	for _, value := range roomMap {
		udpAddr, err := net.ResolveUDPAddr("udp", value.BSList[0])

		//udp连接
		conn, err = net.DialUDP("udp", nil, udpAddr)
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
		proto.SessionId = info.SessionId

		fmt.Println("send login msg to server:" + value.BSList[0])
		conn.Write(proto.ToBytes())
		break
	}

	go processUDPRead(conn)
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

		_, err := conn.Read(allbuf[0:])
		if err != nil {
			fmt.Println("err:" + err.Error())
			continue
		}

		var msgtype int16
		utils.BytesToNum(allbuf[4:8], &msgtype)

		if msgtype == common.MSG_RET_LOGIN {
			proto := new(gtprotocol.RetLoginProtocol)
			proto.Parse(allbuf)

			if proto.Result == 1 {
				fmt.Println("login to bs server ok")
			} else {
				fmt.Println("login to bs server failed")
			}
		} else if msgtype == common.MSG_DATA_TRANS {
			proto := new(gtprotocol.DataTransProtocol)
			proto.Parse(allbuf)

			fmt.Println("recv msg is ", proto.Data)
		}
	}
}
