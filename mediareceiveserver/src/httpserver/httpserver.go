package httpserver

import (
	"fmt"
	"net/http"
	"strconv"
)

type httpServer struct {
	ip   string
	port int
}

func (pserver *httpServer) HandleFunc(pattern string, handler func(ResponseWriter, *Request)) {
	http.HandleFunc(pattern, handler)
}

func (pserver *httpServer) ListenAndServe() {
	if ip == nil {
		http.ListenAndServe(":"+strconv.Itoa(pserver.port), nil)
	} else {
		http.ListenAndServe(pserver.ip+":"+strconv.Itoa(pserver.port), nil)
	}
}
