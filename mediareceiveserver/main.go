package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"mediasrv"
	"net"
	"net/http"
	"net/url"
	"strconv"
)

// func startTCPConn() {
// 	service := "127.0.0.1:9090"
// 	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
// 	conn, err := net.DialTCP("tcp", nil, tcpAddr)

// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	go handleConnection(conn)
// }

var roomtable map[string]*mediasrv.Room
var roomid int
var portid int = 20000

var sip string
var sport int

func genID() int {
	roomid++
	return roomid
}

func getPort() int {
	portid++
	return portid
}

func getCmd(rw http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL.String())
	ip := req.URL.Query().Get("ip")
	port, _ := strconv.Atoi(req.URL.Query().Get("port"))
	fmt.Println("http://" + ip + ":" + strconv.Itoa(port) + "/get")
	resp, err := http.Get("http://" + ip + ":" + strconv.Itoa(port) + "/get")
	defer resp.Body.Close()
	if err != nil {
		// handle error
		fmt.Println(err.Error())
		io.WriteString(rw, "{\"error\":\"http error\"}")
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	// srvtype, _ := strconv.Atoi(req.PostFormValue("type"))
	// server := ServerInfo{srvtype, ip, port}
	// receiveServerArray = append(receiveServerArray, server)
	// fmt.Println("add new server ", ip, ":", port)
	// io.WriteString(w, "add successfully\n")
	// ip := req.URL.Query().Get("ip")
	// port := strconv.Atoi(req.URL.Query().Get("port"))
	// type
	//id := req.URL.Query().Get("id")
	id := strconv.Itoa(genID())
	room := mediasrv.NewRoom(id, sip, getPort(), string(body))
	roomtable[id] = room
	room.Start()

	b, _ := json.Marshal(room)
	retdata := string(b)
	io.WriteString(rw, retdata)
}

func listCmd(rw http.ResponseWriter, req *http.Request) {
	if len(roomtable) > 0 {
		b, _ := json.Marshal(roomtable)
		//json.Encoder.Encode("v")
		io.WriteString(rw, string(b))
	} else {
		io.WriteString(rw, "no room on this server")
	}
}

func startHTTPServer() {
	http.HandleFunc("/get", getCmd)
	http.HandleFunc("/list", listCmd)
	http.ListenAndServe(":4040", nil)
}

func registerServer() {
	resp, err := http.PostForm("http://"+sip+":12345/register",
		url.Values{"ip": {sip}, "port": {"4040"}, "type": {"0"}})

	if err != nil {
		// handle error
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	fmt.Println(string(body))
}

var c chan int

func main() {
	c := make(chan int)
	roomtable = make(map[string]*mediasrv.Room)
	pip := flag.String("ip", "192.168.96.124", "ip address")
	pport := flag.Int("port", 20001, "port")
	flag.Parse()
	sip = *pip
	sport = *pport
	fmt.Println("ip:", sip)
	fmt.Println("port:", sport)

	go startHTTPServer()
	go registerServer()
	//go startUDPServer()

	_ = <-c
	//_, err = conn.Write([]byte("HEAD / HTTP/1.0\r\n\r\n"))
	//_, err = conn.Read(b) / result, err := ioutil.ReadAll(conn)
}

func handleUDPMessage(conn *net.UDPConn) {
	var buf [20]byte

	n, raddr, err := conn.ReadFromUDP(buf[0:])
	if err != nil {
		return
	}

	fmt.Println("msg is ", string(buf[0:n]))

	//WriteToUDP
	//func (c *UDPConn) WriteToUDP(b []byte, addr *UDPAddr) (int, error)
	_, err = conn.WriteToUDP([]byte("nice to see u:"+string(buf[0:n])), raddr)
	if err != nil {
		fmt.Println("err writetoudp:" + err.Error())
	}

	//go tttest(conn, raddr)
	//checkError(err)
}

func tttest(conn *net.UDPConn, raddr *net.UDPAddr) {
	var content string
	for {
		fmt.Scanln(&content)
		_, errs := conn.WriteToUDP([]byte(content), raddr)
		if errs != nil {
			fmt.Println("err writetoudp:" + errs.Error())
		}
	}
}

func startUDPServer() {
	udpaddr, _ := net.ResolveUDPAddr("udp", ":4040")
	conn, err := net.ListenUDP("udp", udpaddr)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(udpaddr.String())

	for {
		handleUDPMessage(conn)
	}
}
