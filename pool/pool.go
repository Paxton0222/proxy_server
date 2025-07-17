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
		&proxy.HttpProxy{
			Address: "localhost:10808",
		},
		&proxy.HttpProxy{
			Address: "localhost:3129",
		},
		&proxy.SSProxy{
			Address:  "localhost:8388",
			Method:   "aes-256-gcm",
			Password: "1234",
		},
		&proxy.VmessProxy{
			Address:       "localhost:10089",
			Uuid:          "60834c02-6962-44d6-b1f3-993452abc1b0",
			Security:      "auto",
			AlterId:       0,
			TransportType: "tcp",
			TransportPath: "",
		},
		&proxy.VmessProxy{
			Address:       "localhost:10086",
			Uuid:          "60834c02-6962-44d6-b1f3-993452abc1b0",
			Security:      "auto",
			AlterId:       0,
			TransportType: "ws",
			TransportPath: "/vmess",
		},
		&proxy.VlessProxy{
			Address:       "localhost:10088",
			Uuid:          "60834c02-6962-44d6-b1f3-993452abc1b0",
			Flow:          "",
			TransportType: "tcp",
			TransportPath: "",
		},
		&proxy.VlessProxy{
			Address:       "localhost:10087",
			Uuid:          "60834c02-6962-44d6-b1f3-993452abc1b0",
			Flow:          "",
			TransportType: "ws",
			TransportPath: "/vless",
		},
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
