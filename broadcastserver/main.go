package main

import (
	"flag"
	"fmt"
)

var quit chan int

var sip string
var sport int
var scip string
var scport int

func main() {
	quit := make(chan int)

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

	udpServiceInit()
	httpServiceInit()

	_ = <-quit
}
