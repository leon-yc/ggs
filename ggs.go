package ggs

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	//init logger first
	"github.com/leon-yc/ggs/internal/initiator"

	//load balancing
	_ "github.com/leon-yc/ggs/internal/pkg/loadbalancing"
	"github.com/leon-yc/ggs/internal/pkg/runtime"
	"github.com/leon-yc/ggs/pkg/qlog"

	//protocols
	_ "github.com/leon-yc/ggs/internal/client/grpc"
	_ "github.com/leon-yc/ggs/internal/client/rest"
	_ "github.com/leon-yc/ggs/internal/server/ginhttp"
	_ "github.com/leon-yc/ggs/internal/server/grpc"

	//routers
	"github.com/leon-yc/ggs/internal/core/common"
	"github.com/leon-yc/ggs/internal/core/config"
	"github.com/leon-yc/ggs/internal/core/handler"
	"github.com/leon-yc/ggs/internal/core/registry"

	//router
	_ "github.com/leon-yc/ggs/internal/core/router/servicecomb"
	//control panel
	_ "github.com/leon-yc/ggs/internal/control/archaius"
	// registry
	"github.com/leon-yc/ggs/internal/core/metadata"
	_ "github.com/leon-yc/ggs/internal/core/registry/consul"
	_ "github.com/leon-yc/ggs/internal/core/registry/file"
	_ "github.com/leon-yc/ggs/internal/core/registry/servicecenter"
	"github.com/leon-yc/ggs/internal/core/server"

	//trace
	_ "github.com/leon-yc/ggs/internal/core/tracing/jaeger"
	// prometheus reporter for circuit breaker metrics
	_ "github.com/leon-yc/ggs/third_party/forked/afex/hystrix-go/hystrix/reporter"
	// aes package handles security related plugins
	_ "github.com/leon-yc/ggs/internal/security/plugins/aes"
	_ "github.com/leon-yc/ggs/internal/security/plugins/plain"

	//config centers
	_ "github.com/go-chassis/go-chassis-config/configcenter"

	//set GOMAXPROCS
	_ "go.uber.org/automaxprocs"

	"github.com/gin-gonic/gin"
	"github.com/go-chassis/go-archaius"
	"github.com/leon-yc/ggs/pkg/metrics"
	"github.com/leon-yc/ggs/third_party/forked/jpillora/overseer"
)

var (
	egn    *engine
	overss overseer.State
)

func init() {
	egn = &engine{}
}

//RegisterSchema Register a API service to specific server by name.
func RegisterSchema(serverName string, structPtr interface{}, opts ...server.RegisterOption) {
	egn.registerSchema(serverName, structPtr, opts...)
}

//Gin return a *gin.Engine that you can register route with gin
func Gin(opts ...server.RegisterOption) (*gin.Engine, error) {
	return egn.gin(opts...)
}

//setDefaultConsumerChains your custom chain map for Consumer,if there is no config, this default chain will take affect
func setDefaultConsumerChains(c map[string]string) {
	egn.DefaultConsumerChainNames = c
}

//setDefaultProviderChains set your custom chain map for Provider,if there is no config, this default chain will take affect
func setDefaultProviderChains(c map[string]string) {
	egn.DefaultProviderChainNames = c
}

//Run bring up the service, it waits for os signal,and shutdown gracefully.
//it support graceful restart default, you can disable it by using ggs.DisableGracefulRestart() as options.
func Run(options ...RunOption) error {
	opts := getRunOpts(options...)
	err := egn.start()
	if err != nil {
		qlog.Error("run engine failed:" + err.Error())
		return err
	}
	if !config.GetRegistratorDisable() {
		//Register instance after Server started
		if err := registry.DoRegister(); err != nil {
			qlog.Error("register instance failed:" + err.Error())
			return err
		}
	}
	waitingSignal(opts)
	return nil
}

func waitingSignal(opts RunOptions) {
	//Graceful shutdown
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGILL, syscall.SIGTRAP, syscall.SIGABRT)
	overseerC := overss.GracefulShutdown
	if overseerC == nil {
		overseerC = make(chan bool)
	}
	isGraceRestart := false
	select {
	case s := <-c:
		qlog.Info("got os signal " + s.String())
	case <-overseerC:
		qlog.Info("got graceful shutdown signal from overseer for restart")
		isGraceRestart = true
	case err := <-server.ErrRuntime:
		qlog.Info("got server error " + err.Error())
	}

	if !config.GetRegistratorDisable() && !isGraceRestart {
		if err := server.UnRegistrySelfInstances(); err != nil {
			qlog.Warnf("servers failed to unregister: %s", err)
		}
	}

	if runtime.InsideDocker {
		const WaitTime = 5
		qlog.Infof("[in docker]sleep %d seconds before graceful shutdown", WaitTime)
		time.Sleep(time.Second * WaitTime)
	}

	for name, s := range server.GetServers() {
		qlog.Info("stopping " + name + " server...")
		err := s.Stop()
		if err != nil {
			qlog.Warnf("servers failed to stop: %s", err)
		}
		qlog.Info(name + " server stop success")
	}

	if archaius.GetBool("ggs.metrics.autometrics.enabled", false) && !isGraceRestart {
		metrics.DeAutoRegistryMetrics()
	}

	if opts.exitCb != nil {
		opts.exitCb()
	}

	qlog.Info("ggs server gracefully shutdown")
}

//Init prepare the ggs framework runtime
func Init(options ...InitOption) error {
	if egn.DefaultConsumerChainNames == nil {
		defaultChain := strings.Join([]string{
			handler.MetricsConsumer,
			handler.RatelimiterConsumer,
			handler.BizkeeperConsumer,
			handler.Loadbalance,
			handler.TracingConsumer,
			handler.Transport,
		}, ",")
		egn.DefaultConsumerChainNames = map[string]string{
			common.DefaultKey: defaultChain,
		}
	}
	if egn.DefaultProviderChainNames == nil {
		defaultChain := strings.Join([]string{
			handler.MetricsProvider,
			handler.RatelimiterProvider,
			handler.LogProvider,
			handler.TracingProvider,
		}, ",")
		egn.DefaultProviderChainNames = map[string]string{
			common.DefaultKey: defaultChain,
		}
	}
	if err := egn.initialize(options...); err != nil {
		qlog.Info("init ggs fail:", err)
		return err
	}
	qlog.Infof("init ggs success, version is %s", metadata.SdkVersion)
	return nil
}

//GraceFork supports graceful restart for rest service by master/slave processes.
func GraceFork(main func()) {
	addrs, err := initiator.ParseListenAddresses()
	if err != nil {
		panic(err)
	}
	if len(addrs) == 0 {
		panic(fmt.Errorf("listen address not parsed from config file"))
	}

	qlog.Infof("listen addresses:%v", addrs)

	prog := func(state overseer.State) {
		qlog.Tracef("got overseer state: %+v", state)
		overss = state
		main()
	}

	overseer.Run(overseer.Config{
		Program:          prog,
		Addresses:        addrs,
		TerminateTimeout: time.Second * 15,
		Debug:            false,
	})
}
