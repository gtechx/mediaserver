package main

import (
	"../common"
	"../common/error"
	"../common/protocol"
	"../common/room"
	"encoding/json"
	"fmt"
	"gtnet"
	"io/ioutil"
	"net/http"
	"runtime"
	"strconv"
	"utils"
)

var clientServer *gtnet.GTUdpServer
var transServer *gtnet.GTUdpServer

var recvLoginChan chan *gtnet.GTUDPPacket
var recvTransChan chan *gtnet.GTUDPPacket

func udpServiceInit() {
	recvLoginChan = make(chan *gtnet.GTUDPPacket, 1024)
	recvTransChan = make(chan *gtnet.GTUDPPacket, 1024)

	startUDPServer()
	startRecvProcess()
	startTransRecvProcess()
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

func startRecvProcess() {
	var numCPU = runtime.NumCPU()
	fmt.Println("startRecvProcess")

	for i := 0; i < numCPU; i++ {
		go func() {
			for packet := range recvLoginChan {
				pack := packet
				msgtype := gtprotocol.GetMsgType(pack.Buff)
				// pack := packet
				// proto := new(gtprotocol.Protocol)
				// proto.Parse(pack.Buff)

				roomname := gtprotocol.GetRoomName(pack.Buff)

				//fmt.Println("process msgtype:", msgtype)
				//fmt.Println("room name:", roomname+"222")

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
				msgtype := gtprotocol.GetMsgType(pack.Buff)
				// pack := packet
				// proto := new(gtprotocol.Protocol)
				// proto.Parse(pack.Buff)

				if msgtype != common.MSG_RS_CONN {
					roomname := gtprotocol.GetRoomName(pack.Buff)

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
