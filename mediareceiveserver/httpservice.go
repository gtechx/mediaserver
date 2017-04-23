package main

import (
	"../common"
	"../common/error"
	"../common/helper/http"
	"../common/protocol"
	"../common/room"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"utils"
)

var roomMap map[string]*gtroom.RSRoom
var bsMap map[string]*net.UDPConn

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

func httpServiceInit() {
	roomMap = make(map[string]*gtroom.RSRoom)
	bsMap = make(map[string]*net.UDPConn)

	go startHTTPServer()
	registerServer()
}

func startHTTPServer() {
	http.HandleFunc("/create", createCmd)
	http.HandleFunc("/list", listCmd)
	http.ListenAndServe(":4040", nil)
}

func registerServer() {
	resp, err := http.Get("http://" + scip + ":" + utils.IntToStr(scport) + "/register?httpport=4040&servertype=rs&ip=" + sip + "&port=" + utils.IntToStr(sport))

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
		ipstr := server.Ip + ":" + utils.IntToStr(server.Port-1)
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

		bsMap[server.Ip+":"+utils.IntToStr(server.Port)] = conn
	}
}

func createCmd(rw http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL.String())
	ip := req.URL.Query().Get("ip")
	port := req.URL.Query().Get("port")
	httpport := req.URL.Query().Get("httpport")
	roomname := req.URL.Query().Get("roomname")

	bsconn, ok := bsMap[ip+":"+port]

	if !ok {
		httphelper.WriteError(rw, 9, "not connect to bs server")
		return
	}

	fmt.Println("http://" + ip + ":" + httpport + "/create?" + "roomname=" + roomname)
	resp, err := http.Get("http://" + ip + ":" + httpport + "/create?" + "roomname=" + roomname)
	defer resp.Body.Close()
	if err != nil {
		// handle error
		fmt.Println(err.Error())
		httphelper.WriteError(rw, 9, "http error")
		return
	}
	body, err := ioutil.ReadAll(resp.Body)

	gterr := new(gterror.Error)
	json.Unmarshal(body, gterr)

	if gterr.ErrorCode != 0 {
		httphelper.WriteError(rw, 3, "serve error")
		return
	}

	rsroom := gtroom.NewRSRoom(roomname)
	rsroom.AddBS(bsconn)
	roomMap[roomname] = rsroom

	fmt.Println("create room success:roomname=" + roomname + " ip=" + ip + " port=" + port)

	httphelper.WriteError(rw, 0, "ok")
}

func listCmd(rw http.ResponseWriter, req *http.Request) {
	if len(roomMap) > 0 {
		b, _ := json.Marshal(roomMap)
		//json.Encoder.Encode("v")
		io.WriteString(rw, string(b))
	} else {
		httphelper.WriteError(rw, 3, "no room on this server")
	}
}
