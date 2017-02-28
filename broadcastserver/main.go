package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

var roomtable map[string]*Room
var roomid int

func main() {
	registerServer()

	for {

	}
}

func registerServer() {
	resp, err := http.PostForm("http://192.168.96.124:12345/register",
		url.Values{"ip": {"192.168.96.123"}, "port": {"3030"}, "type": {"1"}})

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
	http.HandleFunc("get", getCmd)
	http.ListenAndServe(":3031", nil)
}

func genID() int {
	roomid++
	return roomid
}

func getCmd(rw http.ResponseWriter, req *http.Request) {
	id := strconv.Itoa(genID())
	room := NewRoom(id)
	roomtable[id] = room

	b, _ := json.Marshal(roomtable)
	retdata := string(b)
	io.WriteString(rw, retdata)
}
