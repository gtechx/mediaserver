package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
)

type ServerInfo struct {
	Servertype int
	Ip         string
	Port       int
}

var receiveServerArray []ServerInfo
var broadcastServerArray []ServerInfo

// hello world, the web server
func listServer(w http.ResponseWriter, req *http.Request) {
	//io.WriteString(w, "hello, world!\n")
	if len(receiveServerArray) > 0 {
		//retdata := "" + strconv.Itoa(receiveServerArray[0].servertype) + receiveServerArray[0].ip
		b, _ := json.Marshal(receiveServerArray)
		retdata := "{ \"rs\":" + string(b)

		if len(broadcastServerArray) > 0 {
			retdata = retdata + ","
			b, _ = json.Marshal(broadcastServerArray)
			retdata = retdata + "\"bs\":" + string(b)
		}

		retdata = retdata + "}"

		io.WriteString(w, retdata)
	} else {
		io.WriteString(w, "no servers\n")
	}
}

func registerServer(w http.ResponseWriter, req *http.Request) {
	ip := req.PostFormValue("ip")
	port, _ := strconv.Atoi(req.PostFormValue("port"))
	srvtype, _ := strconv.Atoi(req.PostFormValue("type"))
	server := ServerInfo{srvtype, ip, port}

	if srvtype == 0 {
		receiveServerArray = append(receiveServerArray, server)
		fmt.Println("add new rs server ", ip, ":", port)
		io.WriteString(w, "add successfully\n")
	} else {
		broadcastServerArray = append(broadcastServerArray, server)
		fmt.Println("add new bs server ", ip, ":", port)
		io.WriteString(w, "add successfully\n")
	}

}

func createRoom(rw http.ResponseWriter, req *http.Request) {
	if len(receiveServerArray) <= 0 {
		io.WriteString(rw, "receiveServer not start")
		return
	}

	if len(broadcastServerArray) <= 0 {
		io.WriteString(rw, "broadcastServer not start")
		return
	}

	rsrvinfo := &receiveServerArray[0]
	bsrvinfo := &broadcastServerArray[0]
	resp, err := http.Get("http://" + rsrvinfo.Ip + ":" + strconv.Itoa(rsrvinfo.Port) + "/get?ip=" + bsrvinfo.Ip + "&port=" + strconv.Itoa(bsrvinfo.Port))
	defer resp.Body.Close()
	if err != nil {
		// handle error
		fmt.Println(err.Error())
		io.WriteString(rw, "error")
		return
	}

	body, err := ioutil.ReadAll(resp.Body)

	retdata := string(body)
	//fmt.Println(string(body))
	//io.WriteString(rw, string(body))

	// srvinfo = &broadcastServerArray[0]
	// resp, err = http.Get("http://" + srvinfo.Ip + ":" + strconv.Itoa(srvinfo.Port) + "/get")
	// defer resp.Body.Close()
	// if err != nil {
	// 	// handle error
	// 	fmt.Println(err.Error())
	// 	io.WriteString(rw, "error")
	// 	return
	// }
	// body, err = ioutil.ReadAll(resp.Body)

	// retdata = retdata + string(body)

	fmt.Println(retdata)
	io.WriteString(rw, retdata)
}

func startHTTP() {
	http.HandleFunc("/list", listServer)
	http.HandleFunc("/register", registerServer)
	http.HandleFunc("/create", createRoom)
	http.ListenAndServe(":12345", nil)
}

var c chan int

func main() {
	c := make(chan int)
	go startHTTP()

	//go startTCP()
	//go startUDPCon()

	_ = <-c
}

func handleConnection(conn *net.TCPConn) {
	server := ServerInfo{0, "127.0.0.1", 8000}
	receiveServerArray = append(receiveServerArray, server)
	fmt.Println(len(receiveServerArray))
	for {

	}
}

func startTCP() {
	//creat tcp
	fmt.Println("tcp creating")
	service := ":9090"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	ln, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		// handle error
	}
	for {
		conn, err := ln.AcceptTCP()
		if err != nil {
			// handle error
		}
		fmt.Println(conn.RemoteAddr().String() + " connected")
		go handleConnection(conn)
	}
}

func startUDPCon() {
	//conn, err := net.Dial("udp", "127.0.0.1:4040")
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:4040")

	//udp连接
	conn, err := net.DialUDP("udp", nil, udpAddr)
	//defer conn.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	conn.Write([]byte("Hello world!"))

	go processUDPRead(conn)
	go processUDPWrite(conn)
}

func processUDPRead(conn *net.UDPConn) {
	var content string
	for {
		fmt.Scanln(&content)
		conn.Write([]byte(content))
		fmt.Println("send msg is " + content)
	}

}

func processUDPWrite(conn *net.UDPConn) {
	for {
		var msg [128]byte
		conn.Read(msg[0:])
		fmt.Println("recv msg is ", string(msg[0:]))
	}
}
