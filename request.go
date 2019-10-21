package goyave

import (
	"net/http"

	"github.com/System-Glitch/govalidator"
)

type Request struct {
	httpRequest *http.Request
	rules       govalidator.MapData
	data        map[string]interface{}
}

func newRequest(request *http.Request) *Request {
	// TODO implement newRequest
	return nil
}

func (r *Request) validate() {
	// TODO implement validate
}
