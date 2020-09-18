package grpc

import (
	"context"
	"errors"
	"net"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"

	"github.com/leon-yc/ggs/internal/core/common"
	"github.com/leon-yc/ggs/internal/core/invocation"
	"github.com/leon-yc/ggs/internal/core/registry"
	"github.com/leon-yc/ggs/internal/core/server"
	"github.com/leon-yc/ggs/internal/pkg/runtime"
	"github.com/leon-yc/ggs/internal/pkg/util/iputil"
	"github.com/leon-yc/ggs/pkg/qlog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

//err define
var (
	ErrGRPCSvcDescMissing = errors.New("must use server.WithRPCServiceDesc to set desc")
	ErrGRPCSvcType        = errors.New("must set *grpc.ServiceDesc")
)

//const
const (
	Name = "grpc"
)

func init() {
	server.InstallPlugin(Name, New)
}

//Server is grpc server holder
type Server struct {
	s    *grpc.Server
	opts server.Options
}

//Request2Invocation convert grpc protocol to invocation
func Request2Invocation(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo) *invocation.Invocation {
	md, _ := metadata.FromIncomingContext(ctx)
	sourceServices := md.Get(common.HeaderSourceName)
	var sourceService string
	if len(sourceServices) >= 1 {
		sourceService = sourceServices[0]
	}
	m := make(map[string]string, 0)
	inv := &invocation.Invocation{
		MicroServiceName:   runtime.ServiceName,
		SourceMicroService: sourceService,
		Args:               req,
		Protocol:           "grpc",
		SchemaID:           info.FullMethod,
		OperationID:        info.FullMethod,
		Ctx:                context.WithValue(ctx, common.ContextHeaderKey{}, m),
	}
	// set metadata to Ctx
	for k := range md {
		if vs := md.Get(k); len(vs) >= 1 {
			m[k] = vs[0]
		}
	}
	return inv
}

//Stream2Invocation convert grpc protocol to invocation
func Stream2Invocation(stream grpc.ServerStream, info *grpc.StreamServerInfo) *invocation.Invocation {
	ctx := stream.Context()
	md, _ := metadata.FromIncomingContext(ctx)
	sourceServices := md.Get(common.HeaderSourceName)
	var sourceService string
	if len(sourceServices) >= 1 {
		sourceService = sourceServices[0]
	}
	m := make(map[string]string, 0)
	inv := &invocation.Invocation{
		MicroServiceName:   runtime.ServiceName,
		SourceMicroService: sourceService,
		Protocol:           "grpc",
		SchemaID:           info.FullMethod,
		OperationID:        info.FullMethod,
		Ctx:                context.WithValue(ctx, common.ContextHeaderKey{}, m),
	}
	// set metadata to Ctx
	for k := range md {
		if vs := md.Get(k); len(vs) >= 1 {
			m[k] = vs[0]
		}
	}
	return inv
}

//New create grpc server
func New(opts server.Options) server.ProtocolServer {
	return &Server{
		opts: opts,
	}
}

//Register register grpc services
func (s *Server) Register(schema interface{}, options ...server.RegisterOption) (string, error) {
	var opts server.RegisterOptions
	for _, o := range options {
		o(&opts)
	}

	var unaryInts []grpc.UnaryServerInterceptor
	unaryInts = append(unaryInts, wrapUnaryInterceptor(s.opts))
	unaryInts = append(unaryInts, opts.UnaryInts...)

	var streamInts []grpc.StreamServerInterceptor
	streamInts = append(streamInts, wrapStreamInterceptor(s.opts))
	streamInts = append(streamInts, opts.StreamInts...)

	var grpcOpts []grpc.ServerOption
	grpcOpts = append(grpcOpts, grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(unaryInts...)))
	grpcOpts = append(grpcOpts, grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(streamInts...)))

	// new server
	s.s = grpc.NewServer(grpcOpts...)

	if opts.SvcDesc == nil {
		return "", ErrGRPCSvcDescMissing
	}

	// register service
	s.s.RegisterService(opts.SvcDesc, schema)

	// Register reflection service on gRPC server.
	if s.opts.EnableGrpcurl {
		reflection.Register(s.s)
	}
	return "", nil
}

//Start launch the server
func (s *Server) Start() error {
	listen := s.opts.Listen
	if listen == nil {
		l, host, port, lisErr := iputil.StartListener(s.opts.Address, s.opts.TLSConfig)
		if lisErr != nil {
			qlog.Error("listening failed, reason:" + lisErr.Error())
			return lisErr
		}
		registry.InstanceEndpoints[s.opts.ProtocolServerName] = net.JoinHostPort(host, port)
		listen = l
	}

	go func() {
		if err := s.s.Serve(listen); err != nil {
			server.ErrRuntime <- err
		}
	}()
	qlog.Infof("%s server listening on: %s", s.opts.ProtocolServerName, listen.Addr())
	return nil
}

//Stop gracfully shutdown grpc server
func (s *Server) Stop() error {
	stopped := make(chan struct{})
	go func() {
		s.s.GracefulStop()
		close(stopped)
	}()

	t := time.NewTimer(10 * time.Second)
	select {
	case <-t.C:
		s.s.Stop()
	case <-stopped:
		t.Stop()
	}
	return nil
}

//String return server name
func (s *Server) String() string {
	return Name
}
