package main

import (
	"log"
	"net/http"
	"proxy/protocol"
)

func startHTTPProxy(addr string) {
	connProtocol := protocol.ConnectProtocol{}
	directProtocol := protocol.DirectProtocol{}

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodConnect {
			connProtocol.Handle(w, r)
		} else {
			directProtocol.Handle(w, r)
		}
	}

	server := &http.Server{
		Addr:    addr,
		Handler: http.HandlerFunc(handler),
	}

	log.Printf("HTTP Proxy 啟動於 %s\n", addr)
	log.Fatal(server.ListenAndServe())
}

func main() {
	go startHTTPProxy(":8080")

	// 防止主線程退出
	select {}
}
