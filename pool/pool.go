package pool

import (
	"io"
	"math/rand"
	"net"
	"net/http"
	"proxy/proxy"
	"time"
)

type Pool struct{}

func (p *Pool) Handle(clientConn net.Conn, r *http.Request) {
	pool := [...]proxy.Proxy{
		//&proxy.HttpProxy{
		//	Address: "localhost:10808",
		//},
		//&proxy.HttpProxy{
		//	Address: "localhost:3129",
		//},
		//&proxy.SSProxy{
		//	Address:  "localhost:8388",
		//	Method:   "aes-256-gcm",
		//	Password: "1234",
		//},
		//&proxy.VmessProxy{
		//	Address:       "localhost",
		//	Port:          "10089",
		//	Uuid:          "60834c02-6962-44d6-b1f3-993452abc1b0",
		//	Security:      "auto",
		//	AlterId:       0,
		//	TransportType: "tcp",
		//	TransportPath: "",
		//},
		//&proxy.VmessProxy{
		//	Address:       "localhost",
		//	Port:          "10086",
		//	Uuid:          "60834c02-6962-44d6-b1f3-993452abc1b0",
		//	Security:      "auto",
		//	AlterId:       0,
		//	TransportType: "ws",
		//	TransportPath: "/vmess",
		//},
		//&proxy.VlessProxy{
		//	Address:       "localhost",
		//	Port:          "10088",
		//	Uuid:          "60834c02-6962-44d6-b1f3-993452abc1b0",
		//	Flow:          "",
		//	TransportType: "tcp",
		//	TransportPath: "",
		//},
		//&proxy.VlessProxy{
		//	Address:       "localhost",
		//	Port:          "10087",
		//	Uuid:          "60834c02-6962-44d6-b1f3-993452abc1b0",
		//	Flow:          "",
		//	TransportType: "ws",
		//	TransportPath: "/vless",
		//},

		//&proxy.VlessProxy{
		//	Address:          "cis.visa.com",
		//	Port:             "80",
		//	Uuid:             "60834c02-6962-44d6-b1f3-993452abc1b0",
		//	TransportType:    "ws",
		//	TransportHideUrl: "damp-shadow-a1a2.paxton0222.workers.dev",
		//	TransportPath:    "/?ed=2560",
		//},
		//&proxy.VlessProxy{
		//	Address:          "45.80.111.165",
		//	Port:             "2053",
		//	Uuid:             "54f6e78c-b497-4db7-ba48-38c4cf81d5ef",
		//	TransportType:    "ws",
		//	TransportHideUrl: "019309B7.XyZ0.PaGeS.dEV",
		//	TransportPath:    "/",
		//	TlsConfig: option.OutboundTLSOptions{
		//		Enabled:    true,
		//		ServerName: "019309B7.XyZ0.PaGeS.dEV",
		//		Insecure:   false,
		//	},
		//},
		//&proxy.VlessProxy{
		//	Address:          "14.102.229.213",
		//	Port:             "8880",
		//	Uuid:             "fab7bf9c-ddb9-4563-8a04-fb01ce6c0fbf",
		//	TransportType:    "ws",
		//	TransportHideUrl: "jp.laoyoutiao.link",
		//	TransportPath:    "/",
		//},
		//&proxy.VmessProxy{
		//	Address:          "35.212.178.40",
		//	Port:             "8880",
		//	Uuid:             "482c7152-b91b-4081-b1fc-5a0cf13c6635",
		//	AlterId:          0,
		//	Security:         "auto",
		//	TransportType:    "ws",
		//	TransportHideUrl: "",
		//	TransportPath:    "482c7152-b91b-4081-b1fc-5a0cf13c6635-vm",
		//},
		//&proxy.VlessProxy{
		//	Address:          "190.93.246.246",
		//	Port:             "8443",
		//	Uuid:             "b5441b0d-2147-4898-8a6a-9b2c87f58382",
		//	TransportType:    "ws",
		//	TransportHideUrl: "bitget1.asdasd.click",
		//	TransportPath:    "/",
		//	TlsConfig: option.OutboundTLSOptions{
		//		Enabled:    true,
		//		DisableSNI: false,
		//		ServerName: "bitget1.asdasd.click",
		//		Insecure:   true,
		//	},
		//},
	}
	poolLength := len(pool)
	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(poolLength)
	pool[index].Proxy(clientConn, r)
}

func (p *Pool) Direct(w http.ResponseWriter, r *http.Request) {
	if r.URL.Scheme == "http" {
		resp, err := http.DefaultTransport.RoundTrip(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
		}
		defer resp.Body.Close()

		for k, v := range resp.Header {
			for _, val := range v {
				w.Header().Add(k, val)
			}
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	} else {
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
}
