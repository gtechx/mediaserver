package main

import (
	"flag"
	"fmt"
	//"logs"
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

var sip string
var sport int
var scip string
var scport int

var quit chan bool

func main() {
	quit := make(chan bool)

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
	// logs.SetLogger(logs.AdapterFile, `{"filename":"test.log"}`)
	// logs.Debug("ip:", sip)
	// logs.Info("port:", sport)
	// logs.Warn("ip:", sip)
	// logs.Error("ip:", sip)
	// logs.Critical("ip:", sip)

	udpServiceInit()
	httpServiceInit()

	_ = <-quit
}
