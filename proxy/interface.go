package proxy

import "net/http"

type Proxy interface {
	Proxy(w http.ResponseWriter, r *http.Request)
	Direct(w http.ResponseWriter, r *http.Request)
}
