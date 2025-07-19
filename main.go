package main

import (
	"log"
	"net/http"
	"proxy/pool"
	"time"
)

func startHTTPProxy(addr string) {
	//proxies := []*pool.Node{
	//	{
	//		Host: "localhost:10808",
	//		ProxyServer: &proxy.HttpProxy{
	//			Address: "localhost:10808",
	//			Ssl:     false,
	//		},
	//	},
	//	{
	//		Host: "localhost:3129",
	//		ProxyServer: &proxy.HttpProxy{
	//			Address: "localhost:3129",
	//			Ssl:     true,
	//		},
	//	},
	//	{
	//		Host: "localhost:8388",
	//		ProxyServer: &proxy.SSProxy{
	//			Address:  "localhost:8388",
	//			Method:   "aes-256-gcm",
	//			Password: "1234",
	//		},
	//	},
	//	{
	//		Host: "localhost:10089",
	//		ProxyServer: &proxy.VmessProxy{
	//			Address:       "localhost",
	//			Port:          "10089",
	//			Uuid:          "60834c02-6962-44d6-b1f3-993452abc1b0",
	//			Security:      "auto",
	//			AlterId:       0,
	//			TransportType: "tcp",
	//			TransportPath: "",
	//		},
	//	},
	//	{
	//		Host: "localhost:10086",
	//		ProxyServer: &proxy.VmessProxy{
	//			Address:       "localhost",
	//			Port:          "10086",
	//			Uuid:          "60834c02-6962-44d6-b1f3-993452abc1b0",
	//			Security:      "auto",
	//			AlterId:       0,
	//			TransportType: "ws",
	//			TransportPath: "/vmess",
	//		},
	//	},
	//	{
	//		Host: "localhost:10088",
	//		ProxyServer: &proxy.VlessProxy{
	//			Address:       "localhost",
	//			Port:          "10088",
	//			Uuid:          "60834c02-6962-44d6-b1f3-993452abc1b0",
	//			Flow:          "",
	//			TransportType: "tcp",
	//			TransportPath: "",
	//		},
	//	},
	//	{
	//		Host: "localhost:10087",
	//		ProxyServer: &proxy.VlessProxy{
	//			Address:       "localhost",
	//			Port:          "10087",
	//			Uuid:          "60834c02-6962-44d6-b1f3-993452abc1b0",
	//			Flow:          "",
	//			TransportType: "ws",
	//			TransportPath: "/vless",
	//		},
	//	},
	//}

	proxies := pool.LoadProxyConfigFromFile("proxy.txt")
	duration, _ := time.ParseDuration("5m")
	proxyPool := pool.NewPool(proxies)
	go proxyPool.StartHealthCheck(duration, 50)

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
