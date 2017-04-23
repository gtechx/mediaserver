package main

import (
	"../common/helper/http"
	"../common/room"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"utils"
)

var roomMap map[string]*gtroom.BSRoom

func httpServiceInit() {
	roomMap = make(map[string]*gtroom.BSRoom)

	go startHTTPServer()
	registerServer()
}

func registerServer() {
	resp, err := http.Get("http://" + scip + ":" + utils.IntToStr(scport) + "/register?httpport=3030&servertype=bs&ip=" + sip + "&port=" + utils.IntToStr(sport))

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

	httphelper.WriteError(rw, 0, "ok")
}

func listCmd(rw http.ResponseWriter, req *http.Request) {
	if len(roomMap) > 0 {
		b, _ := json.Marshal(roomMap)
		//json.Encoder.Encode("v")
		io.WriteString(rw, string(b))
	} else {
		httphelper.WriteError(rw, 6, "no room on this server")
	}
}
