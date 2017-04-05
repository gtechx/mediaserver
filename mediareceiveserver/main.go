package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"gtnet"
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
var scip string
var scport int

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
	room := mediasrv.NewRoom(id, sip, sport, string(body), scip, scport)
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
	resp, err := http.PostForm("http://"+scip+":"+strconv.Itoa(scport)+"/register",
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

var quit chan bool
var recvChan chan *gtnet.GTUDPPacket

func main() {
	recvChan = make(chan *gtnet.GTUDPPacket, 1024)
	quit := make(chan bool)
	roomtable = make(map[string]*mediasrv.Room)
	lip := flag.String("ip", "192.168.96.124", "ip address")
	lport := flag.Int("port", 20001, "port")

	pip := flag.String("scip", "192.168.96.124", "server center ip address")
	pport := flag.Int("scport", 12345, "server center http port")

	flag.Parse()
	sip = *lip
	sport = *lport
	scip = *pip
	scport = *pport
	fmt.Println("ip:", sip)
	fmt.Println("port:", sport)

	go startHTTPServer()
	go registerServer()
	go startUDPServer()
	startRecvProcess()

	_ = <-quit
	//_, err = conn.Write([]byte("HEAD / HTTP/1.0\r\n\r\n"))
	//_, err = conn.Read(b) / result, err := ioutil.ReadAll(conn)
}

func startRecvProcess() {
	var numCPU = runtime.NumCPU()

	for i := 0; i < numCPU; i++ {
		go func() {
			for packet := range recvChan {
				if g.OnPreSend != nil {
					g.OnPreSend(packet)
				}

				num, err := g.conn.Send(packet.buff, packet.raddr)
				if err != nil {
					fmt.Println("err Send:" + err.Error())
					if g.OnError != nil {
						g.OnError(1, "Send error:"+err.Error())
					}
					return
				}

				if g.onPostSend != nil {
					g.onPostSend(packet, num)
				}
			}
		}()
	}
}

func onRecv(packet *gtnet.GTUDPPacket) {
	recvChan <- packet
}

func startUDPServer() {
	server := gtnet.NewUdpServer(sip, sport)
	server.OnRecv = onRecv
	err := server.Start()

	if err != nil {
		fmt.Println("Server start error:" + err.Error())
	}
}
