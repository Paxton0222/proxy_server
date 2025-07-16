package proxy

import (
	"bufio"
	"github.com/sagernet/sing-vmess/vless"
	"github.com/sagernet/sing/common/logger"
	"github.com/sagernet/sing/common/metadata"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

type VlessProxy struct {
	Address string
	Uuid    string
	Flow    string
}

func (v *VlessProxy) Proxy(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		v.connect(w, r)
	} else {
		v.direct(w, r)
	}
}

func (v *VlessProxy) direct(w http.ResponseWriter, r *http.Request) {
	targetHost := r.URL.Host
	if targetHost == "" {
		targetHost = r.Host
	}
	if _, _, err := net.SplitHostPort(targetHost); err != nil {
		// 若沒有 port，預設 HTTP 為 80
		targetHost = net.JoinHostPort(targetHost, "80")
	}

	// 建立 vless client
	client, err := vless.NewClient(v.Uuid, v.Flow, logger.NOP())
	if err != nil {
		http.Error(w, "VLess client error", http.StatusBadGateway)
		log.Println("VLess client error:", err)
		return
	}

	// 與 vless server 建立 TCP 連線
	upstreamConn, err := net.Dial("tcp", v.Address)
	if err != nil {
		http.Error(w, "Failed to connect to upstream", http.StatusBadGateway)
		log.Println("Dial upstream error:", err)
		return
	}
	defer upstreamConn.Close()

	// 建立 vless stream
	host, port, _ := net.SplitHostPort(targetHost)
	target := metadata.ParseSocksaddrHostPortStr(host, port)
	vlessConn, err := client.DialConn(upstreamConn, target)
	if err != nil {
		http.Error(w, "VLess handshake failed", http.StatusBadGateway)
		log.Println("VLess DialConn error:", err)
		return
	}
	defer vlessConn.Close()

	// 轉發 HTTP 請求原文
	if err := r.Write(vlessConn); err != nil {
		http.Error(w, "Failed to write request", http.StatusBadGateway)
		log.Println("Failed to write HTTP request:", err)
		return
	}

	// 讀取 vless 回應，寫回 client
	resp, err := http.ReadResponse(bufio.NewReader(vlessConn), r)
	if err != nil {
		http.Error(w, "Failed to read response", http.StatusBadGateway)
		log.Println("ReadResponse error:", err)
		return
	}
	defer resp.Body.Close()

	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)

	log.Printf("Client -> Proxy (current) -> %s (vless) -> %s (target)", v.Address, targetHost)
}

func (v *VlessProxy) connect(w http.ResponseWriter, r *http.Request) {
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

	clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	client, err := vless.NewClient(v.Uuid, v.Flow, logger.NOP())
	if err != nil {
		return
	}

	// 連線至 vless
	upstreamConn, err := net.Dial("tcp", v.Address)
	if err != nil {
		return
	}
	defer upstreamConn.Close()

	// 設定 vless 連線至目標
	host, port, err := net.SplitHostPort(r.Host)
	if err != nil {
		return
	}
	vlessConn, err := client.DialConn(
		upstreamConn,
		metadata.ParseSocksaddrHostPortStr(host, port),
	)
	if err != nil {
		return
	}
	defer vlessConn.Close()

	vlessConn.SetDeadline(time.Now().Add(10 * time.Second))

	log.Printf("Client -> Proxy (current) -> %s (vless) -> %s (target)", v.Address, r.Host)

	// client 端與 target 端 互相溝通
	go io.Copy(vlessConn, clientConn)
	io.Copy(clientConn, vlessConn)
}
