package protocol

import (
	"io"
	"net"
	"net/http"
	"proxy/proxy"
)

type ConnectProtocol struct{}

func (p *ConnectProtocol) Handle(w http.ResponseWriter, r *http.Request) {
	//httpsProxy := proxy.HttpsProxy{
	//	Address: "localhost:3129",
	//}
	//httpsProxy.Proxy(w, r)

	//ssProxy := proxy.SSProxy{
	//	Address:  "localhost:8388",
	//	Method:   "aes-256-gcm",
	//	Password: "1234",
	//}
	//ssProxy.Proxy(w, r)

	//vmessProxy := proxy.VmessProxy{
	//	Address:  "localhost:10086",
	//	Uuid:     "60834c02-6962-44d6-b1f3-993452abc1b0",
	//	Security: "auto",
	//	AlterId:  0,
	//}
	//vmessProxy.Proxy(w, r)

	vlessProxy := proxy.VlessProxy{
		Address: "localhost:10087",
		Uuid:    "60834c02-6962-44d6-b1f3-993452abc1b0",
		Flow:    "",
	}
	vlessProxy.Proxy(w, r)
}

func (p *ConnectProtocol) Direct(w http.ResponseWriter, r *http.Request) {

	destConn, err := net.Dial("tcp", r.Host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer destConn.Close()
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer clientConn.Close()
	_, _ = clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	go io.Copy(destConn, clientConn)
	go io.Copy(clientConn, destConn)
}
