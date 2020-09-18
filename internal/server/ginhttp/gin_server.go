package ginhttp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	rt "runtime"
	"strings"
	"sync"

	"github.com/leon-yc/ggs/internal/core/common"
	"github.com/leon-yc/ggs/internal/core/config"
	"github.com/leon-yc/ggs/internal/core/handler"
	"github.com/leon-yc/ggs/internal/core/invocation"
	"github.com/leon-yc/ggs/internal/core/registry"
	"github.com/leon-yc/ggs/internal/core/server"
	"github.com/leon-yc/ggs/internal/pkg/runtime"
	"github.com/leon-yc/ggs/internal/pkg/util/iputil"
	"github.com/leon-yc/ggs/pkg/metrics"
	"github.com/leon-yc/ggs/pkg/qlog"
	"github.com/gin-gonic/gin"
	"github.com/go-chassis/go-archaius"
)

// constants for metric path and name
const (
	//Name is a variable of type string which indicates the protocol being used
	Name               = "rest"
	DefaultMetricPath  = "/metrics"
	DefaultHealthyPath = "/ping"
	MimeFile           = "application/octet-stream"
	MimeMult           = "multipart/form-data"
)

func init() {
	server.InstallPlugin(Name, newGinServer)
}

type ginServer struct {
	microServiceName string
	gs               *gin.Engine
	opts             server.Options
	mux              sync.RWMutex
	exit             chan chan error
	server           *http.Server
}

func newGinServer(opts server.Options) server.ProtocolServer {
	if archaius.GetString("service.environment", "") != "dev" ||
		qlog.GetLevel() != qlog.DebugLevel {
		gin.SetMode(gin.ReleaseMode)
	}
	gs := gin.New()
	gs.Use(wrapHandlerChain(opts))

	if archaius.GetBool("ggs.metrics.enabled", false) {
		metricPath := archaius.GetString("ggs.metrics.apiPath", DefaultMetricPath)
		if !strings.HasPrefix(metricPath, "/") {
			metricPath = "/" + metricPath
		}
		qlog.Info("Enabled metrics API on " + metricPath)
		gs.GET(metricPath, metrics.GinHandleFunc)
	}

	if !archaius.GetBool("ggs.healthy.disabled", false) {
		healthzPath := archaius.GetString("ggs.healthy.apiPath", DefaultHealthyPath)
		if !strings.HasPrefix(healthzPath, "/") {
			healthzPath = "/" + healthzPath
		}
		qlog.Info("Enabled healthy API on " + healthzPath)
		gs.GET(healthzPath, func(c *gin.Context) {
			msg := fmt.Sprintf("Welcome to [%s]%s:%s!", config.MicroserviceDefinition.ServiceDescription.Environment,
				config.MicroserviceDefinition.ServiceDescription.Name,
				config.MicroserviceDefinition.ServiceDescription.Version)
			c.String(http.StatusOK, msg)
		})
	}

	if archaius.GetBool("ggs.pprof.enabled", false) {
		//add pprof
		gs.GET("/debug/pprof/", ginIndex)
		gs.GET("/debug/pprof/cmdline", ginCmdline)
		gs.GET("/debug/pprof/profile", ginProfile)
		gs.GET("/debug/pprof/symbol", ginSymbol)
		gs.GET("/debug/pprof/trace", ginTrace)
		//http.HandleFunc("/debug/pprof/profile", pprof.Profile)
	}

	return &ginServer{
		opts: opts,
		gs:   gs,
	}
}

//wrapHandlerChain wrap business handler with handler chain
func wrapHandlerChain(opts server.Options) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				var stacktrace string
				for i := 1; ; i++ {
					_, f, l, got := rt.Caller(i)
					if !got {
						break
					}

					stacktrace += fmt.Sprintf("%s:%d\n", f, l)
				}
				qlog.WithFields(qlog.Fields{
					"path":  ctx.Request.URL.Path,
					"panic": r,
					"stack": stacktrace,
				}).Error("handle request panic.")
				ctx.String(http.StatusInternalServerError, "server got a panic, plz check log.")
			}
		}()

		c, err := handler.GetChain(common.Provider, opts.ChainName)
		if err != nil {
			qlog.WithError(err).Error("handler chain init err.")
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}
		inv, err := HTTPRequest2Invocation(ctx, "rest", ctx.Request.URL.Path)
		if err != nil {
			qlog.WithError(err).Error("transfer http request to invocation failed.")
			return
		}
		//give inv.Ctx to user handlers, modules may inject headers in handler chain
		c.Next(inv, func(ir *invocation.Response) error {
			if ir.Err != nil {
				ctx.AbortWithStatus(ir.Status)
				return ir.Err
			}
			Invocation2HTTPRequest(inv, ctx)

			// check body size
			if opts.BodyLimit > 0 {
				ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, opts.BodyLimit)
			}

			ctx.Next() //process user's handlers

			ir.Status = ctx.Writer.Status()
			if ir.Status >= http.StatusBadRequest {
				errMsg := ctx.Errors.ByType(gin.ErrorTypePrivate).String()
				if errMsg != "" {
					ir.Err = fmt.Errorf(errMsg)
				} else {
					ir.Err = fmt.Errorf("get err from http handle, get status: %d", ir.Status)
				}
			}
			return ir.Err
		})
	}
}

// HTTPRequest2Invocation convert http request to uniform invocation data format
func HTTPRequest2Invocation(ctx *gin.Context, schema, operation string) (*invocation.Invocation, error) {
	inv := &invocation.Invocation{
		MicroServiceName:   runtime.ServiceName,
		SourceMicroService: common.GetXGGSContext(common.HeaderSourceName, ctx.Request),
		Args:               ctx.Request,
		Protocol:           common.ProtocolRest,
		SchemaID:           schema,
		OperationID:        operation,
		URLPathFormat:      ctx.Request.URL.Path,
		Metadata: map[string]interface{}{
			common.RestMethod: ctx.Request.Method,
		},
	}
	//set headers to Ctx, then user do not  need to consider about protocol in handlers
	m := make(map[string]string, 0)
	inv.Ctx = context.WithValue(context.Background(), common.ContextHeaderKey{}, m)
	for k := range ctx.Request.Header {
		m[k] = ctx.Request.Header.Get(k)
	}
	return inv, nil
}

func (r *ginServer) Register(schema interface{}, opts ...server.RegisterOption) (string, error) {
	qlog.Info("register rest server(gin)")
	return "", nil
}

// Invocation2HTTPRequest convert invocation back to http request, set down all meta data
func Invocation2HTTPRequest(inv *invocation.Invocation, ctx *gin.Context) {
	for k, v := range inv.Metadata {
		ctx.Set(k, v.(string))
	}
	m := common.FromContext(inv.Ctx)
	for k, v := range m {
		ctx.Request.Header.Set(k, v)
	}

	ctx.Request = ctx.Request.WithContext(inv.Ctx)
}

func (r *ginServer) Start() error {
	var err error
	config := r.opts
	r.mux.Lock()
	r.opts.Address = config.Address
	r.mux.Unlock()
	if r.opts.TLSConfig != nil {
		r.server = &http.Server{Addr: config.Address, Handler: r.gs, TLSConfig: r.opts.TLSConfig}
	} else {
		r.server = &http.Server{Addr: config.Address, Handler: r.gs}
	}

	listen := r.opts.Listen
	if listen == nil {
		l, lIP, lPort, err := iputil.StartListener(config.Address, config.TLSConfig)
		if err != nil {
			return fmt.Errorf("failed to start listener: %s", err.Error())
		}
		listen = l
		registry.InstanceEndpoints[config.ProtocolServerName] = net.JoinHostPort(lIP, lPort)
	}

	go func() {
		err = r.server.Serve(listen)
		if err != nil {
			qlog.Error("http server err: " + err.Error())
			server.ErrRuntime <- err
		}

	}()

	qlog.Infof("%s server listening on: %s", r.opts.ProtocolServerName, listen.Addr())
	return nil
}

func (r *ginServer) Stop() error {
	if r.server == nil {
		qlog.Info("http server never started")
		return nil
	}
	//only golang 1.8 support graceful shutdown.
	if err := r.server.Shutdown(context.TODO()); err != nil {
		qlog.Warn("http shutdown error: " + err.Error())
		return err // failure/timeout shutting down the server gracefully
	}
	return nil
}

func (r *ginServer) String() string {
	return Name
}

func (r *ginServer) Engine() interface{} {
	return r.gs
}
