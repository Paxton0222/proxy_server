package proxy

import (
	"io"
	"log"
	"net"
	"net/http"
)

type HttpsProxy struct {
	Address string
}

func (p *HttpsProxy) Proxy(w http.ResponseWriter, r *http.Request) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hj.Hijack()
	if err != nil {
		http.Error(w, "Hijacking failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	defer clientConn.Close()

	// 連線目標主機
	serverConn, err := net.Dial("tcp", r.Host)
	if err != nil {
		clientConn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		return
	}
	defer serverConn.Close()

	// 回應客戶端連線已建立
	clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	log.Printf("Client -> Proxy (current) -> %s (https) -> %s (target)", p.Address, r.Host)

	// 雙向轉發流量
	go io.Copy(serverConn, clientConn)
	io.Copy(clientConn, serverConn)
}
