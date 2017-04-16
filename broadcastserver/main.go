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

var roomMap map[string]*gtroom.BSRoom

var clientServer *gtnet.GTUdpServer
var transServer *gtnet.GTUdpServer
var rsConnList []*net.UDPConn

var recvLoginChan chan *gtnet.GTUDPPacket
var recvTransChan chan *gtnet.GTUDPPacket

var c chan int

var sip string
var sport int
var scip string
var scport int

func writeError(rw http.ResponseWriter, errcode int, errmsg string) {
	io.WriteString(rw, "{\"errorcode\":"+utils.IntToStr(errcode)+", \"error\":\""+errmsg+"\"}")
}

func main() {
	c := make(chan int)
	roomMap = make(map[string]*gtroom.BSRoom)
	recvLoginChan = make(chan *gtnet.GTUDPPacket, 1024)
	recvTransChan = make(chan *gtnet.GTUDPPacket, 1024)

	lip := flag.String("ip", "192.168.1.50", "ip address")
	lport := flag.Int("port", 30001, "port")

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
	startTransRecvProcess()
	registerServer()

	_ = <-c
}

func registerServer() {
	resp, err := http.Get("http://" + scip + ":" + strconv.Itoa(scport) + "/register?httpport=3030&servertype=bs&ip=" + sip + "&port=" + utils.IntToStr(sport))

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

func startHTTPServer() {
	http.HandleFunc("/create", createCmd)
	http.HandleFunc("/list", listCmd)
	http.ListenAndServe(":3030", nil)
}

// type roomInfo struct {
// 	id   string `json:"id1"`
// 	ip   string `json:"ip1"`
// 	port int    `json:"port1"`
// }

func createCmd(rw http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL.String())
	roomname := req.URL.Query().Get("roomname")

	bsroom := gtroom.NewBSRoom(roomname)
	roomMap[roomname] = bsroom

	writeError(rw, 0, "ok")
}

func listCmd(rw http.ResponseWriter, req *http.Request) {
	if len(roomMap) > 0 {
		b, _ := json.Marshal(roomMap)
		//json.Encoder.Encode("v")
		io.WriteString(rw, string(b))
	} else {
		writeError(rw, 6, "no room on this server")
	}
}

func processLogin(packet *gtnet.GTUDPPacket, bsroom *gtroom.BSRoom) {
	proto := new(gtprotocol.ReqLoginProtocol)
	proto.Parse(packet.Buff)

	strsid := utils.Uint64ToStr(proto.SessionId)
	roomname := utils.BytesToStr(proto.RoomName[:])
	resp, err := http.Get("http://" + scip + ":" + strconv.Itoa(scport) + "/checklogin?servertype=bs&sessionid=" + strsid + "&roomname=" + roomname)

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

	fmt.Println(string(body))
	var gterr gterror.Error
	json.Unmarshal(body, &gterr)

	if gterr.ErrorCode != 0 {
		retLoginFailMsg(packet)
		return
	}

	//add client to room
	bsroom.AddClient(strsid, packet.Raddr)

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
	fmt.Println("startRecvProcess")

	for i := 0; i < numCPU; i++ {
		go func() {
			for packet := range recvLoginChan {
				pack := packet
				var msgtype int16
				utils.BytesToNum(pack.Buff[4:8], &msgtype)
				// pack := packet
				// proto := new(gtprotocol.Protocol)
				// proto.Parse(pack.Buff)

				index := bytes.IndexByte(pack.Buff[8:40], 0)
				roomname := string(pack.Buff[8 : 8+index])

				fmt.Println("process msgtype:", msgtype)
				fmt.Println("room name:", roomname+"222")

				bsroom, ok := roomMap[roomname]

				if ok {
					if msgtype == common.MSG_REQ_LOGIN {
						go processLogin(pack, bsroom)
					}
				} else {
					fmt.Println("room not on server")
					retLoginFailMsg(pack)
				}
			}
		}()
	}
}

func startTransRecvProcess() {
	var numCPU = runtime.NumCPU()
	fmt.Println("startTransRecvProcess")

	for i := 0; i < numCPU; i++ {
		go func() {
			for packet := range recvTransChan {
				pack := packet
				var msgtype int16
				utils.BytesToNum(pack.Buff[4:8], &msgtype)
				// pack := packet
				// proto := new(gtprotocol.Protocol)
				// proto.Parse(pack.Buff)

				if msgtype != common.MSG_RS_CONN {
					index := bytes.IndexByte(pack.Buff[8:40], 0)
					roomname := string(pack.Buff[8 : 8+index])

					bsroom, ok := roomMap[roomname]

					if ok {
						if msgtype == common.MSG_DATA_TRANS {
							for _, caddr := range bsroom.ClientList {
								pack.Raddr = caddr
								//fmt.Println("send data to client")
								//fmt.Println("data:", pack.Buff)
								clientServer.Send(pack)
							}
						}
					}
				} else {
					fmt.Println("rs server connected:" + pack.Raddr.String())
				}
			}
		}()
	}
}

func onRecv(packet *gtnet.GTUDPPacket) {
	recvLoginChan <- packet
}

func onTransRecv(packet *gtnet.GTUDPPacket) {
	recvTransChan <- packet
}

func startUDPServer() {
	clientServer = gtnet.NewUdpServer(sip, sport)
	clientServer.OnRecv = onRecv
	err := clientServer.Start()

	if err != nil {
		fmt.Println("Server start error:" + err.Error())
	}
	fmt.Println("start client server success")

	transServer = gtnet.NewUdpServer(sip, sport-1)
	transServer.OnRecv = onTransRecv
	err = transServer.Start()

	if err != nil {
		fmt.Println("Server start error:" + err.Error())
	}
	fmt.Println("start trans server success")
}
