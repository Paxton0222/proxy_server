package pool

import (
	"net"
	"net/http"
)

type ProxyPool interface {
	Handle(clientConn net.Conn, r *http.Request)
	Direct(w http.ResponseWriter, r *http.Request)
}
