package main

import (
	"log"
	"net"
	"net/http"
	"proxy/pool"
)

func startHTTPProxy(addr string) {
	httpProtocol := pool.Pool{}

	handler := func(w http.ResponseWriter, r *http.Request) {
		clientConn, err, done := newHijackClientConn(w)
		if done || err != nil {
			return
		}
		defer clientConn.Close()
		httpProtocol.Handle(clientConn, r)
	}

	server := &http.Server{
		Addr:    addr,
		Handler: http.HandlerFunc(handler),
	}

	log.Printf("HTTP Proxy 啟動於 %s\n", addr)
	log.Fatal(server.ListenAndServe())
}

// 劫持客戶端連線
func newHijackClientConn(w http.ResponseWriter) (net.Conn, error, bool) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return nil, nil, true
	}
	clientConn, _, err := hj.Hijack()
	if err != nil {
		http.Error(w, "Hijack failed: "+err.Error(), http.StatusInternalServerError)
		return nil, nil, true
	}
	return clientConn, err, false
}

func main() {
	go startHTTPProxy(":8080")

	// 防止主線程退出
	select {}
}
