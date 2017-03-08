package main

import (
	//"encoding/json"
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"net"
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
	udpAddr, err := net.ResolveUDPAddr("udp", sip+":30001")

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

func processUDPWrite(conn *net.UDPConn) {
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

		databuf := allbuf[13 : 13+datasize]

		fmt.Println("recv msg is ", databuf)
	}
}
