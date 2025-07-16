package proxy

import (
	"bufio"
	"github.com/shadowsocks/go-shadowsocks2/core"
	"github.com/shadowsocks/go-shadowsocks2/socks"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

type SSProxy struct {
	Address  string
	Method   string
	Password string
}

func (s *SSProxy) Proxy(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		s.connect(w, r)
	} else {
		s.direct(w, r)
	}
}

func (s *SSProxy) direct(w http.ResponseWriter, r *http.Request) {
	cipher, err := core.PickCipher(s.Method, nil, s.Password)
	if err != nil {
		http.Error(w, "cipher error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	ssConn, err := core.Dial("tcp", s.Address, cipher)
	if err != nil {
		http.Error(w, "ss dial error: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer ssConn.Close()
	ssConn.SetDeadline(time.Now().Add(15 * time.Second))

	// 處理 Host 加 port
	host := r.Host
	if _, _, err := net.SplitHostPort(host); err != nil {
		host += ":80"
	}

	// 傳送地址封包
	_, err = ssConn.Write(socks.ParseAddr(host))
	if err != nil {
		http.Error(w, "write target to ss failed: "+err.Error(), http.StatusBadGateway)
		return
	}

	// 清除 URI
	r.RequestURI = ""

	// 寫入 request 給 ss server
	err = r.Write(ssConn)
	if err != nil {
		http.Error(w, "send request via ss failed: "+err.Error(), http.StatusBadGateway)
		return
	}

	// 回應
	resp, err := http.ReadResponse(bufio.NewReader(ssConn), r)
	if err != nil {
		http.Error(w, "read response from ss failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	log.Printf("Client -> Proxy (current) -> %s (ss) -> %s (target)", s.Address, r.Host)
	io.Copy(w, resp.Body)
}

func (s *SSProxy) connect(w http.ResponseWriter, r *http.Request) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hj.Hijack()
	if err != nil {
		http.Error(w, "Hijack failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer clientConn.Close()

	cipher, err := core.PickCipher(s.Method, nil, s.Password)
	if err != nil {
		return
	}
	ssConn, err := core.Dial("tcp", s.Address, cipher)
	if err != nil {
		return
	}
	defer ssConn.Close()

	_, err = ssConn.Write(socks.ParseAddr(r.Host))
	if err != nil {
		log.Println("failed to write target addr to ssConn:", err)
		return
	}

	clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	log.Printf("Client -> Proxy (current) -> %s (ss) -> %s (target)", s.Address, r.Host)

	go io.Copy(ssConn, clientConn)
	io.Copy(clientConn, ssConn)
}
