package protocol

import "net/http"

type Protocol interface {
	Handle(w http.ResponseWriter, r *http.Request)
}
