package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

var roomtable map[string]*Room
var roomid int
var portid int = 30000
var c chan int

var sip string
var sport int

func main() {
	c := make(chan int)
	roomtable = make(map[string]*Room)

	pip := flag.String("ip", "192.168.96.124", "ip address")
	pport := flag.Int("port", 20001, "port")
	flag.Parse()
	sip = *pip
	sport = *pport
	fmt.Println("ip:", sip)
	fmt.Println("port:", sport)

	go startHTTPServer()
	go registerServer()

	_ = <-c
}

func registerServer() {
	resp, err := http.PostForm("http://"+sip+":12345/register",
		url.Values{"ip": {sip}, "port": {"3030"}, "type": {"1"}})

	if err != nil {
		// handle error
		fmt.Println(err.Error())
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	fmt.Println(string(body))
}

func startHTTPServer() {
	http.HandleFunc("/get", getCmd)
	http.HandleFunc("/list", listCmd)
	http.ListenAndServe(":3030", nil)
}

func genID() int {
	roomid++
	return roomid
}

func getPort() int {
	portid++
	return portid
}

// type roomInfo struct {
// 	id   string `json:"id1"`
// 	ip   string `json:"ip1"`
// 	port int    `json:"port1"`
// }

func getCmd(rw http.ResponseWriter, req *http.Request) {
	id := strconv.Itoa(genID())
	room := NewRoom(id, sip, getPort())
	roomtable[id] = room
	room.Start()

	fmt.Println(room.Id, room.Ip, room.Port)
	// var rinfo roomInfo
	// rinfo.id = room.id
	// rinfo.ip = room.ip
	// rinfo.port = room.port
	//rinfo := roomInfo{room.id, room.ip, room.port}
	b, err := json.Marshal(room)
	if err != nil {
		fmt.Println("json err:" + err.Error())
		io.WriteString(rw, "json error")
		return
	}
	fmt.Println("room data:" + string(b))
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
