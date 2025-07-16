package protocol

import (
	"io"
	"net/http"
	"proxy/proxy"
)

type DirectProtocol struct{}

func (p *DirectProtocol) Handle(w http.ResponseWriter, r *http.Request) {
	//httpProxy := proxy.HttpProxy{
	//	Address: "localhost:10808",
	//}
	//httpProxy.Proxy(w, r)

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

func (p *DirectProtocol) Direct(w http.ResponseWriter, r *http.Request) {
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
}
