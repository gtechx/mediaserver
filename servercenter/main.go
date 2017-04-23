package main

import (
	"flag"
	"fmt"
)

var quit chan int

func main() {
	quit := make(chan int)

	pip := flag.String("ip", "192.168.1.50", "ip address")
	pport := flag.Int("port", 20001, "port")
	flag.Parse()
	sip = *pip
	sport = *pport
	fmt.Println("ip:", sip)
	fmt.Println("port:", sport)

	httpServiceInit()

	_ = <-quit
}

// func handleConnection(conn *net.TCPConn) {
// 	server := ServerInfo{0, "127.0.0.1", 8000}
// 	receiveServerArray = append(receiveServerArray, server)
// 	fmt.Println(len(receiveServerArray))
// 	for {

// 	}
// }

// func startTCP() {
// 	//creat tcp
// 	fmt.Println("tcp creating")
// 	service := ":9090"
// 	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
// 	ln, err := net.ListenTCP("tcp", tcpAddr)
// 	if err != nil {
// 		// handle error
// 	}
// 	for {
// 		conn, err := ln.AcceptTCP()
// 		if err != nil {
// 			// handle error
// 		}
// 		fmt.Println(conn.RemoteAddr().String() + " connected")
// 		go handleConnection(conn)
// 	}
// }

// func startUDPCon() {
// 	//conn, err := net.Dial("udp", "127.0.0.1:4040")
// 	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:4040")

// 	//udp连接
// 	conn, err := net.DialUDP("udp", nil, udpAddr)
// 	//defer conn.Close()
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	conn.Write([]byte("Hello world!"))

// 	go processUDPRead(conn)
// 	go processUDPWrite(conn)
// }

// func processUDPRead(conn *net.UDPConn) {
// 	var content string
// 	for {
// 		fmt.Scanln(&content)
// 		conn.Write([]byte(content))
// 		fmt.Println("send msg is " + content)
// 	}

// }

// func processUDPWrite(conn *net.UDPConn) {
// 	for {
// 		var msg [128]byte
// 		conn.Read(msg[0:])
// 		fmt.Println("recv msg is ", string(msg[0:]))
// 	}
// }
