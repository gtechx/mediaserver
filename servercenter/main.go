package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"time"
)

type client struct {
	Ip   string       `json:"ip"`
	Port int          `json:"port"`
	conn *net.UDPConn `json:"-"`
}

type room struct {
	Id       string                 `json:"id"`
	Ip       string                 `json:"ip"`
	Port     int                    `json:"port"`
	conn     *net.UDPConn           `json:"-"`
	iclients map[int64]*net.UDPAddr `json:"-"`
	Clients  map[string]*client     `json:"subroom"`
}

type roominfo struct {
	RoomType    string `json:"type"` //主播，自由
	HasPassword int    `json:"haspassword"`
	password    string `json:"-"`
	IsPublic    int    `json:"ispublic"`
	PRoom       *room  `json:"room"`
	sessionId   string `json:"-"`
}

type userInfo struct {
	account  string
	password string
}

var publicRooms []roominfo
var privateRooms []roominfo
var roomMaps map[string]*roominfo

var sessionIdMaps map[string]*userInfo

type ServerInfo struct {
	Servertype int
	Ip         string
	Port       int
}

var receiveServerArray []ServerInfo
var broadcastServerArray []ServerInfo

var sip string
var sport int

func checkSessionId(sessionid string) bool {
	if _, ok := sessionIdMaps[sessionid]; ok {
		return true
	}

	return false
}

func checkAccountAndPassword(account string, password string) bool {
	intaccount, _ := strconv.Atoi(account)
	//intpassword, _ := strconv.Atoi(password)

	if intaccount > 1000 && intaccount < 10000 {
		return true
	}

	return false
}

// hello world, the web server
func listServers(rw http.ResponseWriter, req *http.Request) {
	//io.WriteString(w, "hello, world!\n")
	if len(receiveServerArray) > 0 {
		//retdata := "" + strconv.Itoa(receiveServerArray[0].servertype) + receiveServerArray[0].ip
		b, _ := json.Marshal(receiveServerArray)
		retdata := "{ \"rs\":" + string(b)

		if len(broadcastServerArray) > 0 {
			retdata = retdata + ","
			b, _ = json.Marshal(broadcastServerArray)
			retdata = retdata + "\"bs\":" + string(b)
		}

		retdata = retdata + "}"

		io.WriteString(rw, retdata)
	} else {
		io.WriteString(rw, "{\"error\":\"no servers\"}")
	}
}

func listRooms(rw http.ResponseWriter, req *http.Request) {
	sessionid := req.URL.Query().Get("sessionid")

	if checkSessionId(sessionid) == false {
		io.WriteString(rw, "{\"errorcode\":2, \"error\":\"account or password is not right\"}")
		return
	}

	if len(publicRooms) > 0 {
		b, _ := json.Marshal(publicRooms)
		io.WriteString(rw, string(b))
	} else {
		io.WriteString(rw, "{\"errorcode\":3, \"error\":\"no rooms\"}")
	}
}

func registerServer(w http.ResponseWriter, req *http.Request) {
	ip := req.PostFormValue("ip")
	port, _ := strconv.Atoi(req.PostFormValue("port"))
	srvtype, _ := strconv.Atoi(req.PostFormValue("type"))
	server := ServerInfo{srvtype, ip, port}

	if srvtype == 0 {
		receiveServerArray = append(receiveServerArray, server)
		fmt.Println("add new rs server ", ip, ":", port)
		io.WriteString(w, "add successfully\n")
	} else {
		broadcastServerArray = append(broadcastServerArray, server)
		fmt.Println("add new bs server ", ip, ":", port)
		io.WriteString(w, "add successfully\n")
	}

}

func createRoom(rw http.ResponseWriter, req *http.Request) {
	sessionid := req.URL.Query().Get("sessionid")

	if checkSessionId(sessionid) == false {
		io.WriteString(rw, "{\"errorcode\":2, \"error\":\"account or password is not right\"}")
		return
	}

	if len(receiveServerArray) <= 0 {
		io.WriteString(rw, "{\"errorcode\":3, \"error\":\"receiveServer not start\"}")
		return
	}

	if len(broadcastServerArray) <= 0 {
		io.WriteString(rw, "{\"errorcode\":3, \"error\":\"broadcastServer not start\"}")
		return
	}

	rsrvinfo := &receiveServerArray[0]
	bsrvinfo := &broadcastServerArray[0]
	resp, err := http.Get("http://" + rsrvinfo.Ip + ":" + strconv.Itoa(rsrvinfo.Port) + "/get?ip=" + bsrvinfo.Ip + "&port=" + strconv.Itoa(bsrvinfo.Port))
	defer resp.Body.Close()
	if err != nil {
		// handle error
		fmt.Println(err.Error())
		io.WriteString(rw, "error")
		return
	}

	body, err := ioutil.ReadAll(resp.Body)

	roomtype := req.URL.Query().Get("type")
	if roomtype == "" {
		fmt.Println("roomtype is blank")
		roomtype = "ziyou"
	} else {
		fmt.Println("roomtype:", roomtype)
	}

	password := req.URL.Query().Get("password")
	haspassword := 1
	if password == "" {
		haspassword = 0
	}

	pubtype := req.URL.Query().Get("ispublic")
	ispublic := 0
	if pubtype == "" {
		ispublic = 1
	}

	var croom room
	json.Unmarshal(body, &croom)
	fmt.Println(croom)
	newroom := roominfo{roomtype, haspassword, password, ispublic, &croom, sessionid}

	if ispublic == 1 {
		publicRooms = append(publicRooms, newroom)
	} else {
		privateRooms = append(privateRooms, newroom)
	}

	roomMaps[croom.Id] = &newroom

	retdata := string(body)
	//fmt.Println(string(body))
	//io.WriteString(rw, string(body))

	// srvinfo = &broadcastServerArray[0]
	// resp, err = http.Get("http://" + srvinfo.Ip + ":" + strconv.Itoa(srvinfo.Port) + "/get")
	// defer resp.Body.Close()
	// if err != nil {
	// 	// handle error
	// 	fmt.Println(err.Error())
	// 	io.WriteString(rw, "error")
	// 	return
	// }
	// body, err = ioutil.ReadAll(resp.Body)

	// retdata = retdata + string(body)

	fmt.Println(retdata)
	io.WriteString(rw, retdata)
}

func checkLogin(rw http.ResponseWriter, req *http.Request) {
	id := req.URL.Query().Get("id")
	srvtype := req.URL.Query().Get("srvtype")
	sessionid := req.URL.Query().Get("sessionid")
	proominfo := roomMaps[id]
	uinfo := sessionIdMaps[sessionid]
	useraccount := uinfo.account
	//userpassword := uinfo.password

	if checkSessionId(sessionid) == false {
		io.WriteString(rw, "{\"errorcode\":2, \"error\":\"account or password is not right\"}")
		return
	}

	if srvtype == "rs" {
		if proominfo.RoomType == "zhubo" {
			if sessionid == proominfo.sessionId {
				fmt.Println("zhubo:" + useraccount + " logined success")
				io.WriteString(rw, "{\"ok\":\"login success\"}")
			} else {
				fmt.Println("zhubo:" + useraccount + " login failed")
				io.WriteString(rw, "{\"errorcode\":2,\"error\":\"account or password is not right\"}")
			}
		} else {
			if proominfo.HasPassword == 1 {
				password := req.URL.Query().Get("password")

				if password == proominfo.password {
					fmt.Println("user:" + useraccount + " logined in ziyou rs room:" + proominfo.PRoom.Id + " success")
					io.WriteString(rw, "{\"ok\":\"login success\"}")
				} else {
					fmt.Println("user:" + useraccount + " logined in ziyou rs room:" + proominfo.PRoom.Id + " failed")
					io.WriteString(rw, "{\"errorcode\":2,\"error\":\"password is not right\"}")
				}
			} else {
				fmt.Println("user:" + useraccount + " logined in ziyou rs room:" + proominfo.PRoom.Id + " success, without password")
				io.WriteString(rw, "{\"ok\":\"login success\"}")
			}
		}
	} else if srvtype == "bs" {
		if proominfo.HasPassword == 1 {
			password := req.URL.Query().Get("password")

			if password == proominfo.password {
				fmt.Println("user:" + useraccount + " logined in bs room:" + proominfo.PRoom.Id + " success")
				io.WriteString(rw, "{\"ok\":\"login success\"}")
			} else {
				fmt.Println("user:" + useraccount + " logined in bs room:" + proominfo.PRoom.Id + " failed")
				io.WriteString(rw, "{\"errorcode\":2,\"error\":\"password is not right\"}")
			}
		} else {
			fmt.Println("user:" + useraccount + " logined in bs room:" + proominfo.PRoom.Id + " success, without password")
			io.WriteString(rw, "{\"ok\":\"login success\"}")
		}
	} else {
		fmt.Println("user:" + useraccount + " try to login in room:" + proominfo.PRoom.Id + " with error srvtype:" + srvtype + "(rs/bs)")
		io.WriteString(rw, "{\"errorcode\":1,\"error\":\"login failed\"}")
	}
}

func userLogin(rw http.ResponseWriter, req *http.Request) {
	useraccount := req.URL.Query().Get("useraccount")
	userpassword := req.URL.Query().Get("userpassword")

	if checkAccountAndPassword(useraccount, userpassword) == false {
		io.WriteString(rw, "{\"errorcode\":2, \"error\":\"account or password is not right\"}")
		return
	}

	sessionid := strconv.FormatInt(time.Now().UnixNano(), 10)
	accinfo := userInfo{useraccount, userpassword}
	sessionIdMaps[sessionid] = &accinfo

	io.WriteString(rw, "{\"uid\":"+sessionid+"}")
}

func startHTTP() {
	http.HandleFunc("/listservers", listServers)
	http.HandleFunc("/listrooms", listRooms)
	http.HandleFunc("/register", registerServer)
	http.HandleFunc("/create", createRoom)
	http.HandleFunc("/login", userLogin)
	http.HandleFunc("/checklogin", checkLogin)
	http.ListenAndServe(":12345", nil)
}

var c chan int

func main() {
	roomMaps = make(map[string]*roominfo)
	c := make(chan int)
	pip := flag.String("ip", "192.168.96.124", "ip address")
	pport := flag.Int("port", 20001, "port")
	flag.Parse()
	sip = *pip
	sport = *pport
	fmt.Println("ip:", sip)
	fmt.Println("port:", sport)
	go startHTTP()

	//go startTCP()
	//go startUDPCon()

	_ = <-c
}

func handleConnection(conn *net.TCPConn) {
	server := ServerInfo{0, "127.0.0.1", 8000}
	receiveServerArray = append(receiveServerArray, server)
	fmt.Println(len(receiveServerArray))
	for {

	}
}

func startTCP() {
	//creat tcp
	fmt.Println("tcp creating")
	service := ":9090"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	ln, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		// handle error
	}
	for {
		conn, err := ln.AcceptTCP()
		if err != nil {
			// handle error
		}
		fmt.Println(conn.RemoteAddr().String() + " connected")
		go handleConnection(conn)
	}
}

func startUDPCon() {
	//conn, err := net.Dial("udp", "127.0.0.1:4040")
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:4040")

	//udp连接
	conn, err := net.DialUDP("udp", nil, udpAddr)
	//defer conn.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	conn.Write([]byte("Hello world!"))

	go processUDPRead(conn)
	go processUDPWrite(conn)
}

func processUDPRead(conn *net.UDPConn) {
	var content string
	for {
		fmt.Scanln(&content)
		conn.Write([]byte(content))
		fmt.Println("send msg is " + content)
	}

}

func processUDPWrite(conn *net.UDPConn) {
	for {
		var msg [128]byte
		conn.Read(msg[0:])
		fmt.Println("recv msg is ", string(msg[0:]))
	}
}
