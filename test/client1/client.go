package main

import (
	//"encoding/json"
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
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

func startUDPCon() {
	// resp, err := http.Get("http://192.168.96.124:12345/create")
	// defer resp.Body.Close()
	// if err != nil {
	// 	// handle error
	// 	fmt.Println(err.Error())
	// 	io.WriteString(rw, "{\"error\":\"http error\"}")
	// 	return
	// }
	// body, err := ioutil.ReadAll(resp.Body)

	// var rinfo roomInfo
	// json.Unmarshal(body, &rinfo)

	// b, _ := json.Marshal(rinfo)
	// fmt.Println("room info:" + string(b))

	//conn, err := net.Dial("udp", "127.0.0.1:4040")
	udpAddr, err := net.ResolveUDPAddr("udp", sip+":20001")

	//udp连接
	conn, err := net.DialUDP("udp", nil, udpAddr)
	//defer conn.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	buf := make([]byte, 13)
	buf[12] = 0
	conn.Write(buf)
	//conn.Write([]byte("Hello world!"))

	go processUDPRead(conn)
	go processUDPWrite(conn)
}

func processUDPRead(conn *net.UDPConn) {
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
			var uid int64
			var dtype byte

			uid = 1012
			dtype = 2
			datasize = int32(num)

			bytesBuffer := bytes.NewBuffer([]byte{})
			binary.Write(bytesBuffer, binary.LittleEndian, datasize)
			sendbuf := bytesBuffer.Bytes()

			bytesBuffer = bytes.NewBuffer([]byte{})
			binary.Write(bytesBuffer, binary.LittleEndian, uid)
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

func processUDPWrite(conn *net.UDPConn) {
	for {
		var msg [2048]byte
		conn.ReadFromUDP(msg[0:])
		fmt.Println("recv msg is ", string(msg[0:]))
	}
}
