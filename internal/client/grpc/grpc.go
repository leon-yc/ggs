package grpc

import (
	"context"
	"errors"
	"time"

	"github.com/leon-yc/ggs/internal/core/client"
	"github.com/leon-yc/ggs/internal/core/common"
	"github.com/leon-yc/ggs/internal/core/invocation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

func init() {
	client.InstallPlugin("grpc", New)
}

const errPrefix = "grpc client: "

//Client is grpc client holder
type Client struct {
	c       *grpc.ClientConn
	opts    client.Options
	service string
	timeout time.Duration
}

//New create new grpc client
func New(opts client.Options) (client.ProtocolClient, error) {
	conn, err := newClientConn(opts)
	if err != nil {
		err = errors.New(errPrefix + err.Error())
		return nil, err
	}

	return &Client{
		c:       conn,
		timeout: opts.Timeout,
		service: opts.Service,
		opts:    opts,
	}, nil
}
func newClientConn(opts client.Options) (*grpc.ClientConn, error) {
	var conn *grpc.ClientConn
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()
	if opts.TLSConfig == nil {
		conn, err = grpc.DialContext(ctx, opts.Endpoint, grpc.WithInsecure())
	} else {
		conn, err = grpc.DialContext(ctx, opts.Endpoint,
			grpc.WithTransportCredentials(credentials.NewTLS(opts.TLSConfig)))
	}
	return conn, err
}

//TransformContext will deliver header in chassis context key to grpc context key
func TransformContext(ctx context.Context) context.Context {
	m := common.FromContext(ctx)
	md := metadata.New(m)
	return metadata.NewOutgoingContext(ctx, md)
}

//Call remote server
func (c *Client) Call(ctx context.Context, addr string, inv *invocation.Invocation, rsp interface{}) (err error) {
	ctx = TransformContext(ctx)
	if inv.IsStream {
		streamDesc, ok := (inv.StreamDesc).(*grpc.StreamDesc)
		if !ok {
			return errors.New("stream require *grpc.StreamDesc")
		}
		var stream grpc.ClientStream
		stream, err = c.c.NewStream(ctx, streamDesc, "/"+inv.SchemaID+"/"+inv.OperationID)
		if stream != nil {
			if resp, ok := rsp.(*grpc.ClientStream); ok {
				*resp = stream
			}
		}
	} else {
		ctx, cancel := context.WithTimeout(ctx, c.timeout)
		err = c.c.Invoke(ctx, "/"+inv.SchemaID+"/"+inv.OperationID, inv.Args, rsp)
		cancel()
	}
	return
}

//String return name
func (c *Client) String() string {
	return "grpc"
}

// Close close conn
func (c *Client) Close() error {
	return c.c.Close()
}

// ReloadConfigs reload configs for timeout and tls
func (c *Client) ReloadConfigs(opts client.Options) {
	newOpts := client.EqualOpts(c.opts, opts)
	if newOpts.TLSConfig != c.opts.TLSConfig {
		conn, err := newClientConn(opts)
		if err == nil && conn != nil {
			if c.c != nil {
				c.c.Close()
			}
			c.c = conn
		}
	}

	c.opts = newOpts
	c.timeout = newOpts.Timeout
}

// GetOptions method return opts
func (c *Client) GetOptions() client.Options {
	return c.opts
}
