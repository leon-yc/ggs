package server

import (
	"crypto/tls"
	"net"

	"github.com/leon-yc/ggs/internal/core/provider"
	"google.golang.org/grpc"
)

//Options is the options for service initiating
type Options struct {
	Address            string
	Listen             net.Listener
	ProtocolServerName string
	ChainName          string
	Provider           provider.Provider
	TLSConfig          *tls.Config
	BodyLimit          int64
	EnableGrpcurl      bool // 开启grpcurl
}

//RegisterOptions is options when you register a schema to chassis
type RegisterOptions struct {
	SchemaID   string
	ServerName string
	// grpc 相关
	SvcDesc    *grpc.ServiceDesc
	UnaryInts  []grpc.UnaryServerInterceptor
	StreamInts []grpc.StreamServerInterceptor
}

//RegisterOption is option when you register a schema to ggs
type RegisterOption func(*RegisterOptions)

// newRegisterOptions is for updating options
func NewRegisterOptions(options ...RegisterOption) RegisterOptions {
	opts := RegisterOptions{}
	for _, o := range options {
		o(&opts)
	}
	return opts
}
