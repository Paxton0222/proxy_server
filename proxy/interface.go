package proxy

import (
	"net"
	"net/http"
)

type Proxy interface {
	Proxy(clientConn net.Conn, r *http.Request)
	Request(r *http.Request) (*http.Response, error)
}
