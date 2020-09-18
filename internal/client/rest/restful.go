package rest

import (
	"io"
	"net/http"
	"sync"

	"github.com/leon-yc/ggs/internal/core/client"
)

//NewRequest is a function which creates new request
func NewRequest(method, urlStr string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, urlStr, body)
}

// NewResponse is creating the object of response
func NewResponse() *http.Response {
	resp := &http.Response{
		Header: http.Header{},
	}
	return resp
}

//Client is a struct
type Client struct {
	c    *http.Client
	opts client.Options
	mu   sync.Mutex // protects following
}
