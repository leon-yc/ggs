package ggs

import (
	"fmt"
	"net"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/go-chassis/go-archaius"
	"github.com/leon-yc/ggs/internal/bootstrap"
	"github.com/leon-yc/ggs/internal/configcenter"
	"github.com/leon-yc/ggs/internal/control"
	"github.com/leon-yc/ggs/internal/core/common"
	"github.com/leon-yc/ggs/internal/core/config"
	"github.com/leon-yc/ggs/internal/core/handler"
	"github.com/leon-yc/ggs/internal/core/loadbalancer"
	"github.com/leon-yc/ggs/internal/core/registry"
	"github.com/leon-yc/ggs/internal/core/router"
	"github.com/leon-yc/ggs/internal/core/server"
	"github.com/leon-yc/ggs/internal/core/tracing"
	"github.com/leon-yc/ggs/internal/eventlistener"
	"github.com/leon-yc/ggs/internal/pkg/circuit"
	"github.com/leon-yc/ggs/internal/pkg/runtime"
	"github.com/leon-yc/ggs/internal/pkg/util/fileutil"
	"github.com/leon-yc/ggs/pkg/metrics"
	"github.com/leon-yc/ggs/pkg/qlog"
	"github.com/leon-yc/ggs/third_party/forked/afex/hystrix-go/hystrix"
)

type engine struct {
	version     string
	schemas     []*Schema
	mu          sync.Mutex
	Initialized bool

	DefaultConsumerChainNames map[string]string
	DefaultProviderChainNames map[string]string
}

// Schema struct for to represent schema info
type Schema struct {
	serverName string
	schema     interface{}
	opts       []server.RegisterOption
}

func (c *engine) initChains(chainType string) error {
	var defaultChainName = "default"
	var handlerNameMap = map[string]string{defaultChainName: ""}
	switch chainType {
	case common.Provider:
		if providerChainMap := config.GlobalDefinition.Ggs.Handler.Chain.Provider; len(providerChainMap) != 0 {
			if _, ok := providerChainMap[defaultChainName]; !ok {
				providerChainMap[defaultChainName] = c.DefaultProviderChainNames[defaultChainName]
			}
			handlerNameMap = providerChainMap
		} else {
			handlerNameMap = c.DefaultProviderChainNames
		}
	case common.Consumer:
		if consumerChainMap := config.GlobalDefinition.Ggs.Handler.Chain.Consumer; len(consumerChainMap) != 0 {
			if _, ok := consumerChainMap[defaultChainName]; !ok {
				consumerChainMap[defaultChainName] = c.DefaultConsumerChainNames[defaultChainName]
			}
			handlerNameMap = consumerChainMap
		} else {
			handlerNameMap = c.DefaultConsumerChainNames
		}
	}
	qlog.Tracef("init %s's handler map", chainType)
	return handler.CreateChains(chainType, handlerNameMap)
}
func (c *engine) initHandler() error {
	if err := c.initChains(common.Provider); err != nil {
		qlog.Errorf("chain int failed: %s", err)
		return err
	}
	if err := c.initChains(common.Consumer); err != nil {
		qlog.Errorf("chain int failed: %s", err)
		return err
	}
	qlog.Trace("chain init success")
	return nil
}

//Init
func (c *engine) initialize(options ...InitOption) error {
	if c.Initialized {
		return nil
	}

	opts := getInitOpts(options...)
	if opts.configDir != "" {
		qlog.Infof("set config dir to %s", opts.configDir)
		fileutil.SetConfDir(opts.configDir)
	}

	if err := config.Init(); err != nil {
		qlog.Error("failed to initialize conf: " + err.Error())
		return err
	}
	if err := runtime.Init(); err != nil {
		return err
	}

	err := c.initHandler()
	if err != nil {
		qlog.Errorf("handler init failed: %s", err)
		return err
	}

	if err := metrics.Init(); err != nil {
		return err
	}

	listens := make(map[string]net.Listener)
	if overss.Enabled {
		for i := 0; i < len(overss.Addresses); i++ {
			listens[overss.Addresses[i]] = overss.Listeners[i]
		}
	}
	err = server.Init(listens)
	if err != nil {
		return err
	}
	bootstrap.Bootstrap()
	if !archaius.GetBool("ggs.service.registry.disabled", false) {
		err := registry.Enable()
		if err != nil {
			return err
		}
		strategyName := archaius.GetString("ggs.loadbalance.strategy.name", "")
		if err := loadbalancer.Enable(strategyName); err != nil {
			return err
		}
	}

	err = configcenter.Init()
	if err != nil {
		qlog.Warn("lost config server: " + err.Error())
	}
	// router needs get configs from config-center when init
	// so it must init after bootstrap
	if err = router.Init(); err != nil {
		return err
	}
	ctlOpts := control.Options{
		Infra:   config.GlobalDefinition.Panel.Infra,
		Address: config.GlobalDefinition.Panel.Settings["address"],
	}
	if err := control.Init(ctlOpts); err != nil {
		return err
	}

	if !archaius.GetBool("ggs.tracing.disabled", false) {
		if err = tracing.Init(); err != nil {
			return err
		}
	}
	go hystrix.StartReporter()
	circuit.Init()
	eventlistener.Init()
	c.Initialized = true
	return nil
}

func (c *engine) registerSchema(serverName string, structPtr interface{}, opts ...server.RegisterOption) {
	schema := &Schema{
		serverName: serverName,
		schema:     structPtr,
		opts:       opts,
	}
	c.mu.Lock()
	c.schemas = append(c.schemas, schema)
	c.mu.Unlock()
}

func (c *engine) gin(opts ...server.RegisterOption) (*gin.Engine, error) {
	if !c.Initialized {
		return nil, fmt.Errorf("the ggs do not init. please run ggs.Init() first")
	}
	regOpts := server.NewRegisterOptions(opts...)
	serverName := regOpts.ServerName
	if serverName == "" {
		serverName = "rest"
	}
	s, err := server.GetServer(serverName)
	if err != nil {
		return nil, err
	}
	if ginServer, ok := s.(server.GinServer); ok {
		return ginServer.Engine().(*gin.Engine), nil
	}
	return nil, fmt.Errorf("server(%s) is not implemented with gin", serverName)
}

func (c *engine) start() error {
	if !c.Initialized {
		return fmt.Errorf("the ggs do not init. please run ggs.Init() first")
	}

	for _, v := range c.schemas {
		if v == nil {
			continue
		}
		s, err := server.GetServer(v.serverName)
		if err != nil {
			return err
		}
		_, err = s.Register(v.schema, v.opts...)
		if err != nil {
			return err
		}
	}
	err := server.StartServer()
	if err != nil {
		return err
	}
	return nil
}
