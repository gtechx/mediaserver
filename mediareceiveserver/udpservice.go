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

//var bsServer *gtnet.GTUdpServer

var recvChan chan *gtnet.GTUDPPacket

func udpServiceInit() {
	recvChan = make(chan *gtnet.GTUDPPacket, 1024)
	startUDPServer()
	startRecvProcess()
}

func startRecvProcess() {
	var numCPU = runtime.NumCPU()

	for i := 0; i < numCPU; i++ {
		go func() {
			for packet := range recvChan {
				pack := packet
				//var msgtype int16
				//utils.BytesToNum(pack.Buff[4:8], &msgtype)
				msgtype := gtprotocol.GetMsgType(pack.Buff)
				// pack := packet
				// proto := new(gtprotocol.Protocol)
				// proto.Parse(pack.Buff)
				//index := bytes.IndexByte(pack.Buff[8:40], 0)
				roomname := gtprotocol.GetRoomName(pack.Buff) //string(pack.Buff[8 : 8+index])
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
