package protocol

import (
	"net/http"
	"proxy/proxy"
)

type DirectProtocol struct{}

func (p *DirectProtocol) Handle(w http.ResponseWriter, r *http.Request) {
	httpProxy := proxy.HttpProxy{
		Address: "localhost:10808",
	}
	httpProxy.Proxy(w, r)
}
