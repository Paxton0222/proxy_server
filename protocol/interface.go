package protocol

import "net/http"

type Protocol interface {
	Handle(w http.ResponseWriter, r *http.Request)
	Direct(w http.ResponseWriter, r *http.Request)
}
