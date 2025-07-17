package proxy

import (
	"net"
	"net/http"
)

type Proxy interface {
	Proxy(clientConn net.Conn, r *http.Request)
}
