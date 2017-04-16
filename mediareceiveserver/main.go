package main

import (
	"../common"
	"../common/error"
	"../common/protocol"
	"../common/room"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"gtnet"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"utils"
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

var roomMap map[string]*gtroom.RSRoom
var bsMap map[string]*net.UDPConn

var clientServer *gtnet.GTUdpServer
var bsServer *gtnet.GTUdpServer

var sip string
var sport int
var scip string
var scport int

type ServerInfo struct {
	Ip         string
	Port       int
	HttpPort   int
	Servertype string
	ClientNum  int
}

type BSServer struct {
	BSArray []*ServerInfo `json:"bs"`
}

func writeError(rw http.ResponseWriter, errcode int, errmsg string) {
	io.WriteString(rw, "{\"errorcode\":"+utils.IntToStr(errcode)+", \"error\":\""+errmsg+"\"}")
}

func createCmd(rw http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL.String())
	ip := req.URL.Query().Get("ip")
	port := req.URL.Query().Get("port")
	httpport := req.URL.Query().Get("httpport")
	roomname := req.URL.Query().Get("roomname")

	bsconn, ok := bsMap[ip+":"+port]

	if !ok {
		writeError(rw, 9, "not connect to bs server")
		return
	}

	fmt.Println("http://" + ip + ":" + httpport + "/create?" + "roomname=" + roomname)
	resp, err := http.Get("http://" + ip + ":" + httpport + "/create?" + "roomname=" + roomname)
	defer resp.Body.Close()
	if err != nil {
		// handle error
		fmt.Println(err.Error())
		writeError(rw, 9, "http error")
		return
	}
	body, err := ioutil.ReadAll(resp.Body)

	gterr := new(gterror.Error)
	json.Unmarshal(body, gterr)

	if gterr.ErrorCode != 0 {
		writeError(rw, 3, "serve error")
		return
	}

	rsroom := gtroom.NewRSRoom(roomname)
	rsroom.AddBS(bsconn)
	roomMap[roomname] = rsroom

	fmt.Println("create room success:roomname=" + roomname + " ip=" + ip + " port=" + port)

	writeError(rw, 0, "ok")
}

func listCmd(rw http.ResponseWriter, req *http.Request) {
	if len(roomMap) > 0 {
		b, _ := json.Marshal(roomMap)
		//json.Encoder.Encode("v")
		io.WriteString(rw, string(b))
	} else {
		writeError(rw, 3, "no room on this server")
	}
}

func startHTTPServer() {
	http.HandleFunc("/create", createCmd)
	http.HandleFunc("/list", listCmd)
	http.ListenAndServe(":4040", nil)
}

func registerServer() {
	resp, err := http.Get("http://" + scip + ":" + strconv.Itoa(scport) + "/register?httpport=4040&servertype=rs&ip=" + sip + "&port=" + utils.IntToStr(sport))

	if err != nil {
		// handle error
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	bsserver := new(BSServer)
	err = json.Unmarshal(body, bsserver)

	if err != nil {
		fmt.Println(err.Error())
	}
	//fmt.Println(bsserver)
	fmt.Println(string(body))

	for _, server := range bsserver.BSArray {
		//udp连接
		ipstr := server.Ip + ":" + strconv.Itoa(server.Port-1)
		udpAddr, _ := net.ResolveUDPAddr("udp", ipstr)
		conn, err := net.DialUDP("udp", nil, udpAddr)
		if err != nil {
			fmt.Println("connect rs server failed:" + err.Error())
			continue
		}
		fmt.Println("connect rs server ok")
		proto := new(gtprotocol.Protocol)
		proto.DataSize = 0
		proto.MsgType = common.MSG_RS_CONN
		conn.Write(proto.ToBytes())

		bsMap[server.Ip+":"+strconv.Itoa(server.Port)] = conn
	}
}

var quit chan bool
var recvChan chan *gtnet.GTUDPPacket

func main() {
	recvChan = make(chan *gtnet.GTUDPPacket, 1024)
	quit := make(chan bool)
	roomMap = make(map[string]*gtroom.RSRoom)
	bsMap = make(map[string]*net.UDPConn)
	lip := flag.String("ip", "192.168.1.50", "ip address")
	lport := flag.Int("port", 20001, "port")

	pip := flag.String("scip", "192.168.1.50", "server center ip address")
	pport := flag.Int("scport", 12345, "server center http port")

	flag.Parse()
	sip = *lip
	sport = *lport
	scip = *pip
	scport = *pport
	fmt.Println("ip:", sip)
	fmt.Println("port:", sport)

	go startHTTPServer()
	go startUDPServer()
	startRecvProcess()
	registerServer()

	_ = <-quit
	//_, err = conn.Write([]byte("HEAD / HTTP/1.0\r\n\r\n"))
	//_, err = conn.Read(b) / result, err := ioutil.ReadAll(conn)
}

func processLogin(packet *gtnet.GTUDPPacket, rsroom *gtroom.RSRoom) {
	proto := new(gtprotocol.ReqLoginProtocol)
	err := proto.Parse(packet.Buff)

	if err != nil {
		fmt.Println(err.Error())
	}

	strsid := utils.Uint64ToStr(proto.SessionId)
	fmt.Println("login sessionid is ", proto.SessionId)
	roomname := utils.BytesToStr(proto.RoomName[:])
	fmt.Println("http://" + scip + ":" + strconv.Itoa(scport) + "/checklogin?servertype=rs&sessionid=" + strsid + "&roomname=" + roomname)
	resp, err := http.Get("http://" + scip + ":" + strconv.Itoa(scport) + "/checklogin?servertype=rs&sessionid=" + strsid + "&roomname=" + roomname)

	if err != nil {
		fmt.Println(err.Error())
		retLoginFailMsg(packet)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println(err.Error())
		retLoginFailMsg(packet)
		return
	}

	var gterr gterror.Error
	json.Unmarshal(body, &gterr)

	if gterr.ErrorCode != 0 {
		fmt.Println(string(body))
		retLoginFailMsg(packet)
		return
	}

	//add client to room
	rsroom.AddClient(strsid, packet.Raddr)

	//ret login success msg
	retproto := new(gtprotocol.RetLoginProtocol)
	retproto.DataSize = 1
	retproto.MsgType = common.MSG_RET_LOGIN
	retproto.Result = 1

	packet.Buff = retproto.ToBytes()
	clientServer.Send(packet)
}

func retLoginFailMsg(packet *gtnet.GTUDPPacket) {
	retproto := new(gtprotocol.RetLoginProtocol)
	retproto.DataSize = 1
	retproto.MsgType = common.MSG_RET_LOGIN
	retproto.Result = 0

	packet.Buff = retproto.ToBytes()
	clientServer.Send(packet)
}

func startRecvProcess() {
	var numCPU = runtime.NumCPU()

	for i := 0; i < numCPU; i++ {
		go func() {
			for packet := range recvChan {
				pack := packet
				var msgtype int16
				utils.BytesToNum(pack.Buff[4:8], &msgtype)
				// pack := packet
				// proto := new(gtprotocol.Protocol)
				// proto.Parse(pack.Buff)
				index := bytes.IndexByte(pack.Buff[8:40], 0)
				roomname := string(pack.Buff[8 : 8+index])
				//b, _ := json.Marshal(roomMap)
				//fmt.Println(string(b))
				rsroom, ok := roomMap[roomname]

				//fmt.Println("process msgtype:", msgtype)
				//fmt.Println("room name:", roomname+"222")
				if ok {
					if msgtype == common.MSG_REQ_LOGIN {
						go processLogin(pack, rsroom)
					} else if msgtype == common.MSG_DATA_TRANS {
						for _, bsconn := range rsroom.BSList {
							//fmt.Println("send to bs server")
							//fmt.Println("data:", pack.Buff)
							bsconn.Write(pack.Buff)
						}
					}
				} else {
					fmt.Println("room not on server")
					retLoginFailMsg(pack)
				}

			}
		}()
	}
}

func onRecv(packet *gtnet.GTUDPPacket) {
	recvChan <- packet
}

// func onBSRecv(packet *gtnet.GTUDPPacket) {
// 	recvChan <- packet
// }

func startUDPServer() {
	clientServer = gtnet.NewUdpServer(sip, sport)
	clientServer.OnRecv = onRecv
	err := clientServer.Start()

	if err != nil {
		fmt.Println("Server start error:" + err.Error())
	}

	// bsServer = gtnet.NewUdpServer(sip, sport-1)
	// bsServer.OnRecv = onBSRecv
	// err = bsServer.Start()

	// if err != nil {
	// 	fmt.Println("Server start error:" + err.Error())
	// }
}
