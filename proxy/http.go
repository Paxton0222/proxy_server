package proxy

import (
	"io"
	"log"
	"net/http"
	"net/url"
)

type HttpProxy struct {
	Address string
}

func (p *HttpProxy) Proxy(w http.ResponseWriter, r *http.Request) {
	proxyURL, err := url.Parse("http://" + p.Address)
	if err != nil {
		http.Error(w, "Invalid proxy address", http.StatusInternalServerError)
		return
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	client := &http.Client{
		Transport: transport,
	}

	// 清除可能會影響代理的 RequestURI
	r.RequestURI = ""

	resp, err := client.Do(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 複製回應 Header
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(resp.StatusCode)
	log.Printf("Client -> Proxy (current) -> %s (http) -> %s (target)", p.Address, r.Host)
	io.Copy(w, resp.Body)
}

func (p *HttpProxy) Direct(w http.ResponseWriter, r *http.Request) {
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
