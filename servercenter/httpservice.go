package main

import (
	"../common/error"
	"../common/helper/http"
	"../common/room"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
	"utils"
)

var publicRoomMap map[string]*gtroom.SCRoom
var privateRoomMap map[string]*gtroom.SCRoom
var roomMap map[string]*gtroom.SCRoom
var sessionIdMap map[string]string

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

var rsArray []*ServerInfo
var bsArray []*ServerInfo

var sip string
var sport int

func httpServiceInit() {
	publicRoomMap = make(map[string]*gtroom.SCRoom)
	privateRoomMap = make(map[string]*gtroom.SCRoom)
	sessionIdMap = make(map[string]string)
	roomMap = make(map[string]*gtroom.SCRoom)

	go startHTTP()
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

func checkSessionId(sessionid string) bool {
	if _, ok := sessionIdMap[sessionid]; ok {
		return true
	}

	return false
}

func checkAccountAndPassword(account string, password string) bool {
	intaccount, _ := utils.StrToInt(account)
	//intpassword, _ := strconv.Atoi(password)

	if intaccount >= 1000 && intaccount <= 10000 {
		return true
	}

	return false
}

// hello world, the web server
func listServers(rw http.ResponseWriter, req *http.Request) {
	//io.WriteString(w, "hello, world!\n")
	if len(rsArray) > 0 {
		//retdata := "" + strconv.Itoa(receiveServerArray[0].servertype) + receiveServerArray[0].ip
		b, _ := json.Marshal(rsArray)
		retdata := "{ \"rs\":" + string(b)

		if len(bsArray) > 0 {
			retdata = retdata + ","
			b, _ = json.Marshal(bsArray)
			retdata = retdata + "\"bs\":" + string(b)
		}

		retdata = retdata + "}"

		io.WriteString(rw, retdata)
	} else {
		httphelper.WriteError(rw, 9, "no servers")
	}
}

func listRooms(rw http.ResponseWriter, req *http.Request) {
	sessionid := req.URL.Query().Get("sessionid")

	if checkSessionId(sessionid) == false {
		httphelper.WriteError(rw, 2, "account or password is not right")
		return
	}

	fmt.Println("sessionid:" + sessionid + "req listrooms...")
	if len(publicRoomMap) > 0 {
		b, _ := json.Marshal(publicRoomMap)
		io.WriteString(rw, string(b))
	} else {
		fmt.Println("no rooms")
		httphelper.WriteError(rw, 3, "no rooms")
	}
}

func registerServer(w http.ResponseWriter, req *http.Request) {
	ip := req.URL.Query().Get("ip")
	port, _ := utils.StrToInt(req.URL.Query().Get("port"))
	httpport, _ := utils.StrToInt(req.URL.Query().Get("httpport"))
	servertype := req.URL.Query().Get("servertype")
	server := &ServerInfo{ip, port, httpport, servertype, 0}

	if servertype == "rs" {
		rsArray = append(rsArray, server)
		fmt.Println("add new rs server ", ip, ":", port)
		//io.WriteString(w, "add successfully\n")

		if len(bsArray) > 0 {
			bsserver := new(BSServer)
			bsserver.BSArray = bsArray
			//retdata := "" + strconv.Itoa(receiveServerArray[0].servertype) + receiveServerArray[0].ip
			b, _ := json.Marshal(bsserver)
			retdata := "{\"bs\":" + string(b)

			retdata = retdata + "}"
			bsserver1 := new(BSServer)
			err := json.Unmarshal(b, bsserver1)
			if err != nil {
				fmt.Println(err.Error())
			}

			io.WriteString(w, string(b))
		} else {
			io.WriteString(w, "add successfully and no bs server\n")
		}
	} else if servertype == "bs" {
		bsArray = append(bsArray, server)
		fmt.Println("add new bs server ", ip, ":", port)

		io.WriteString(w, "add successfully\n")
	}
}

func createRoom(rw http.ResponseWriter, req *http.Request) {
	sessionid := req.URL.Query().Get("sessionid")
	roomname := sessionIdMap[sessionid]

	if checkSessionId(sessionid) == false {
		httphelper.WriteError(rw, 2, "account or password is not right")
		return
	}

	if _, ok := roomMap[roomname]; ok {
		httphelper.WriteError(rw, 2, "room has exist")
		return
	}

	if len(rsArray) <= 0 {
		httphelper.WriteError(rw, 3, "receiveServer not start")
		return
	}

	if len(bsArray) <= 0 {
		httphelper.WriteError(rw, 3, "broadcastServer not start")
		return
	}

	rsrvinfo := rsArray[0]
	bsrvinfo := bsArray[0]
	resp, err := http.Get("http://" + rsrvinfo.Ip + ":" + utils.IntToStr(rsrvinfo.HttpPort) + "/create?ip=" + bsrvinfo.Ip + "&port=" + utils.IntToStr(bsrvinfo.Port) + "&roomname=" + roomname + "&httpport=" + utils.IntToStr(bsrvinfo.HttpPort))
	defer resp.Body.Close()
	if err != nil {
		// handle error
		fmt.Println(err.Error())
		httphelper.WriteError(rw, 3, "http error")
		return
	}

	body, err := ioutil.ReadAll(resp.Body)

	gterr := new(gterror.Error)
	json.Unmarshal(body, gterr)

	if gterr.ErrorCode != 0 {
		httphelper.WriteError(rw, 3, "serve error")
		return
	}

	roomtype := req.URL.Query().Get("type")
	password := req.URL.Query().Get("password")
	pubtype := req.URL.Query().Get("ispublic")

	if roomtype == "" {
		fmt.Println("roomtype is blank")
		roomtype = "ziyou"
	} else {
		fmt.Println("roomtype:", roomtype)
	}

	haspassword := int8(1)
	if password == "" {
		haspassword = 0
	}

	ispublic := int8(0)
	if pubtype == "" {
		ispublic = 1
	}

	//var croom room
	//json.Unmarshal(body, &croom)
	//fmt.Println(croom)
	scroom := gtroom.NewSCRoom(roomname, ispublic, haspassword, password, roomtype)

	if ispublic == 1 {
		publicRoomMap[roomname] = scroom
	} else {
		privateRoomMap[roomname] = scroom
	}

	roomMap[roomname] = scroom

	scroom.RSList = append(scroom.RSList, rsrvinfo.Ip+":"+utils.IntToStr(rsrvinfo.Port))
	scroom.BSList = append(scroom.BSList, bsrvinfo.Ip+":"+utils.IntToStr(bsrvinfo.Port))

	scbyte, _ := json.Marshal(scroom)
	retdata := string(scbyte)

	fmt.Println(retdata)
	io.WriteString(rw, retdata)
}

func checkLogin(rw http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL.String())
	//id := req.URL.Query().Get("id")
	servertype := req.URL.Query().Get("servertype")
	sessionid := req.URL.Query().Get("sessionid")
	roomname := req.URL.Query().Get("roomname")
	useraccount := sessionIdMap[sessionid]
	scroom, ok := roomMap[roomname]

	fmt.Println("roomname:" + roomname + "55")

	//useraccount := uinfo.account
	//userpassword := uinfo.password

	if !ok {
		httphelper.WriteError(rw, 2, "room not exist")
		return
	}

	if checkSessionId(sessionid) == false {
		httphelper.WriteError(rw, 2, "account or password is not right")
		return
	}

	if servertype == "rs" {
		if scroom.RoomType == "zhubo" {
			if useraccount == scroom.Name {
				fmt.Println("zhubo:" + useraccount + " logined success")
				httphelper.WriteError(rw, 0, "login success")
			} else {
				fmt.Println("zhubo:" + useraccount + " login failed")
				httphelper.WriteError(rw, 2, "account or password is not right")
			}
		} else {
			if scroom.HasPassword == 1 {
				password := req.URL.Query().Get("password")

				if password == scroom.Password {
					fmt.Println("user:" + useraccount + " logined in ziyou rs room:" + scroom.Name + " success")
					httphelper.WriteError(rw, 0, "login success")
				} else {
					fmt.Println("user:" + useraccount + " logined in ziyou rs room:" + scroom.Name + " failed")
					httphelper.WriteError(rw, 2, "room password is not right")
				}
			} else {
				fmt.Println("user:" + useraccount + " logined in ziyou rs room:" + scroom.Name + " success, without password")
				httphelper.WriteError(rw, 0, "login success")
			}
		}
	} else if servertype == "bs" {
		if scroom.HasPassword == 1 {
			password := req.URL.Query().Get("password")

			if password == scroom.Password {
				fmt.Println("user:" + useraccount + " logined in bs room:" + scroom.Name + " success")
				httphelper.WriteError(rw, 0, "login success")
			} else {
				fmt.Println("user:" + useraccount + " logined in bs room:" + scroom.Name + " failed")
				httphelper.WriteError(rw, 2, "room password is not right")
			}
		} else {
			fmt.Println("user:" + useraccount + " logined in bs room:" + scroom.Name + " success, without password")
			httphelper.WriteError(rw, 0, "login success")
		}
	} else {
		fmt.Println("user:" + useraccount + " try to login in room:" + scroom.Name + " with error srvtype:" + servertype + "(rs/bs)")
		httphelper.WriteError(rw, 1, "login failed")
	}
}

func Md5(text string) string {
	hashMd5 := md5.New()
	io.WriteString(hashMd5, text)
	return fmt.Sprintf("%x", hashMd5.Sum(nil))
}

func userLogin(rw http.ResponseWriter, req *http.Request) {
	useraccount := req.URL.Query().Get("useraccount")
	userpassword := req.URL.Query().Get("userpassword")

	if checkAccountAndPassword(useraccount, userpassword) == false {
		httphelper.WriteError(rw, 2, "account or password is not right")
		return
	}

	intacc, _ := utils.StrToInt64(useraccount)                    // strconv.Atoi(useraccount)
	sessionid := utils.Int64ToStr(time.Now().UnixNano() + intacc) //strconv.FormatInt(time.Now().UnixNano()+int64(intacc), 10)

	// nano := time.Now().UnixNano()
	// rand.Seed(nano)
	// rndNum := rand.Int63()
	// sessionid := Md5(Md5(strconv.FormatInt(nano, 10)) + Md5(strconv.FormatInt(rndNum, 10)))
	// sessionid = sessionid + sessionid

	//accinfo := userInfo{useraccount, userpassword}
	sessionIdMap[sessionid] = useraccount
	fmt.Println("user:" + useraccount + " logined, sessionid=" + sessionid)

	io.WriteString(rw, "{\"sessionid\":"+sessionid+"}")
}
