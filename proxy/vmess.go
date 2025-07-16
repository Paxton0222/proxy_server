package proxy

import (
	"bufio"
	vmess "github.com/sagernet/sing-vmess"
	"github.com/sagernet/sing/common/metadata"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

type VmessProxy struct {
	Address  string
	Uuid     string
	Security string
	AlterId  int
}

func (v *VmessProxy) Proxy(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		v.connect(w, r)
	} else {
		v.direct(w, r)
	}
}

func (v *VmessProxy) direct(w http.ResponseWriter, r *http.Request) {
	targetHost := r.URL.Host
	if targetHost == "" {
		targetHost = r.Host
	}
	if _, _, err := net.SplitHostPort(targetHost); err != nil {
		// 若沒有 port，預設 HTTP 為 80
		targetHost = net.JoinHostPort(targetHost, "80")
	}

	// 建立 VMess client
	client, err := vmess.NewClient(v.Uuid, v.Security, v.AlterId)
	if err != nil {
		http.Error(w, "VMess client error", http.StatusBadGateway)
		log.Println("VMess client error:", err)
		return
	}

	// 與 VMess server 建立 TCP 連線
	upstreamConn, err := net.Dial("tcp", v.Address)
	if err != nil {
		http.Error(w, "Failed to connect to upstream", http.StatusBadGateway)
		log.Println("Dial upstream error:", err)
		return
	}
	defer upstreamConn.Close()

	// 建立 VMess stream
	host, port, _ := net.SplitHostPort(targetHost)
	target := metadata.ParseSocksaddrHostPortStr(host, port)
	vmessConn, err := client.DialConn(upstreamConn, target)
	if err != nil {
		http.Error(w, "VMess handshake failed", http.StatusBadGateway)
		log.Println("VMess DialConn error:", err)
		return
	}
	defer vmessConn.Close()

	// 轉發 HTTP 請求原文
	if err := r.Write(vmessConn); err != nil {
		http.Error(w, "Failed to write request", http.StatusBadGateway)
		log.Println("Failed to write HTTP request:", err)
		return
	}

	// 讀取 VMess 回應，寫回 client
	resp, err := http.ReadResponse(bufio.NewReader(vmessConn), r)
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

	log.Printf("Client -> Proxy (current) -> %s (vmess) -> %s (target)", v.Address, targetHost)
}

func (v *VmessProxy) connect(w http.ResponseWriter, r *http.Request) {
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

	client, err := vmess.NewClient(
		v.Uuid,
		v.Security,
		v.AlterId,
	)
	if err != nil {
		return
	}

	// 連線至 vmess
	upstreamConn, err := net.Dial("tcp", v.Address)
	if err != nil {
		return
	}
	defer upstreamConn.Close()

	// 設定 vmess 連線至目標
	host, port, err := net.SplitHostPort(r.Host)
	if err != nil {
		return
	}
	vmessConn, err := client.DialConn(
		upstreamConn,
		metadata.ParseSocksaddrHostPortStr(host, port),
	)
	if err != nil {
		return
	}
	defer vmessConn.Close()

	vmessConn.SetDeadline(time.Now().Add(10 * time.Second))

	log.Printf("Client -> Proxy (current) -> %s (vmess) -> %s (target)", v.Address, r.Host)

	// client 端與 target 端 互相溝通
	go io.Copy(vmessConn, clientConn)
	io.Copy(clientConn, vmessConn)
}
