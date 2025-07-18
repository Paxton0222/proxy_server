package main

import (
	"github.com/sagernet/sing-box/option"
	"log"
	"net/http"
	"proxy/pool"
	"proxy/proxy"
	"time"
)

func startHTTPProxy(addr string) {
	proxies := []*pool.ProxyNode{
		{
			Host: "localhost:10808",
			ProxyServer: &proxy.HttpProxy{
				Address: "localhost:10808",
				Ssl:     false,
			},
		},
		{
			Host: "localhost:3129",
			ProxyServer: &proxy.HttpProxy{
				Address: "localhost:3129",
				Ssl:     true,
			},
		},
		{
			Host: "localhost:8388",
			ProxyServer: &proxy.SSProxy{
				Address:  "localhost:8388",
				Method:   "aes-256-gcm",
				Password: "1234",
			},
		},
		{
			Host: "localhost:10089",
			ProxyServer: &proxy.VmessProxy{
				Address:       "localhost",
				Port:          "10089",
				Uuid:          "60834c02-6962-44d6-b1f3-993452abc1b0",
				Security:      "auto",
				AlterId:       0,
				TransportType: "tcp",
				TransportPath: "",
			},
		},
		{
			Host: "localhost:10086",
			ProxyServer: &proxy.VmessProxy{
				Address:       "localhost",
				Port:          "10086",
				Uuid:          "60834c02-6962-44d6-b1f3-993452abc1b0",
				Security:      "auto",
				AlterId:       0,
				TransportType: "ws",
				TransportPath: "/vmess",
			},
		},
		{
			Host: "localhost:10088",
			ProxyServer: &proxy.VlessProxy{
				Address:       "localhost",
				Port:          "10088",
				Uuid:          "60834c02-6962-44d6-b1f3-993452abc1b0",
				Flow:          "",
				TransportType: "tcp",
				TransportPath: "",
			},
		},
		{
			Host: "localhost:10087",
			ProxyServer: &proxy.VlessProxy{
				Address:       "localhost",
				Port:          "10087",
				Uuid:          "60834c02-6962-44d6-b1f3-993452abc1b0",
				Flow:          "",
				TransportType: "ws",
				TransportPath: "/vless",
			},
		},
		{
			Host: "damp-shadow-ala2.paxton0222.workers.dev",
			ProxyServer: &proxy.VlessProxy{
				Address:          "cis.visa.com",
				Port:             "80",
				Uuid:             "60834c02-6962-44d6-b1f3-993452abc1b0",
				TransportType:    "ws",
				TransportHideUrl: "damp-shadow-a1a2.paxton0222.workers.dev",
				TransportPath:    "/?ed=2560",
			},
		},
		{
			Host: "45.80.111.165:2053",
			ProxyServer: &proxy.VlessProxy{
				Address:          "45.80.111.165",
				Port:             "2053",
				Uuid:             "54f6e78c-b497-4db7-ba48-38c4cf81d5ef",
				TransportType:    "ws",
				TransportHideUrl: "019309B7.XyZ0.PaGeS.dEV",
				TransportPath:    "/",
				TlsConfig: option.OutboundTLSOptions{
					Enabled:    true,
					ServerName: "019309B7.XyZ0.PaGeS.dEV",
					Insecure:   false,
				},
			},
		},
		{
			Host: "14.102.229.213",
			ProxyServer: &proxy.VlessProxy{
				Address:          "14.102.229.213",
				Port:             "8880",
				Uuid:             "fab7bf9c-ddb9-4563-8a04-fb01ce6c0fbf",
				TransportType:    "ws",
				TransportHideUrl: "jp.laoyoutiao.link",
				TransportPath:    "/",
			},
		},
		{
			Host: "35.212.178.40:8880",
			ProxyServer: &proxy.VmessProxy{
				Address:          "35.212.178.40",
				Port:             "8880",
				Uuid:             "482c7152-b91b-4081-b1fc-5a0cf13c6635",
				AlterId:          0,
				Security:         "auto",
				TransportType:    "ws",
				TransportHideUrl: "",
				TransportPath:    "482c7152-b91b-4081-b1fc-5a0cf13c6635-vm",
			},
		},
		{
			Host: "190.93.246.246:8443",
			ProxyServer: &proxy.VlessProxy{
				Address:          "190.93.246.246",
				Port:             "8443",
				Uuid:             "b5441b0d-2147-4898-8a6a-9b2c87f58382",
				TransportType:    "ws",
				TransportHideUrl: "bitget1.asdasd.click",
				TransportPath:    "/",
				TlsConfig: option.OutboundTLSOptions{
					Enabled:    true,
					DisableSNI: false,
					ServerName: "bitget1.asdasd.click",
					Insecure:   true,
				},
			},
		},
	}
	duration, _ := time.ParseDuration("5m")
	proxyPool := pool.NewPool(proxies)
	go proxyPool.StartHealthCheck(duration, 5)

	server := &http.Server{
		Addr: addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			proxyPool.Handle(w, r)
		}),
	}

	log.Printf("HTTP ProxyServer 啟動於 %s\n", addr)
	log.Fatal(server.ListenAndServe())
}

func main() {
	go startHTTPProxy(":8080")

	// 防止主線程退出
	select {}
}
