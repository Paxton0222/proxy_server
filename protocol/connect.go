package protocol

import (
	"net/http"
	"proxy/proxy"
)

type ConnectProtocol struct{}

func (p *ConnectProtocol) Handle(w http.ResponseWriter, r *http.Request) {
	httpsProxy := proxy.HttpsProxy{
		Address: "localhost:3129",
	}
	httpsProxy.Proxy(w, r)
}
