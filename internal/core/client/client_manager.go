package client

import (
	"fmt"
	"strings"
	"sync"

	"crypto/tls"
	"errors"
	"time"

	"github.com/leon-yc/ggs/internal/core/common"
	"github.com/leon-yc/ggs/internal/core/config"
	"github.com/leon-yc/ggs/internal/core/config/model"
	ggsTLS "github.com/leon-yc/ggs/internal/core/tls"
	"github.com/leon-yc/ggs/pkg/qlog"
)

var clients = make(map[string]ProtocolClient)
var sl sync.RWMutex

//ErrClientNotExist happens if client do not exist
var ErrClientNotExist = errors.New("client not exists")

//DefaultPoolSize is 500
const DefaultPoolSize = 512

//Options is configs for client creation
type Options struct {
	Service   string
	PoolSize  int
	Timeout   time.Duration
	Endpoint  string
	PoolTTL   time.Duration
	TLSConfig *tls.Config
	Failure   map[string]bool
}

// GetFailureMap return failure map
func GetFailureMap(p string) map[string]bool {
	failureList := strings.Split(config.GlobalDefinition.Ggs.Transport.Failure[p], ",")
	failureMap := make(map[string]bool)
	for _, v := range failureList {
		if v == "" {
			continue
		}
		failureMap[v] = true
	}
	return failureMap
}

//GetMaxIdleCon get max idle connection number you defined
//default is 512
func GetMaxIdleCon(p string) int {
	n, ok := config.GetTransportConf().MaxIdlCons[p]
	if !ok {
		return DefaultPoolSize
	}
	return n
}

// CreateClient is for to create client based on protocol and the service name
func CreateClient(protocol, service, endpoint string) (ProtocolClient, error) {
	f, err := GetClientNewFunc(protocol)
	if err != nil {
		qlog.Error(fmt.Sprintf("do not support [%s] client", protocol))
		return nil, err
	}
	tlsConfig, sslConfig, err := ggsTLS.GetTLSConfigByService(service, protocol, common.Consumer)
	if err != nil {
		if !ggsTLS.IsSSLConfigNotExist(err) {
			return nil, err
		}
	} else {
		// client verify target micro service's name in mutual tls
		// remember to set SAN (Subject Alternative Name) as server's micro service name
		// when generating server.csr
		tlsConfig.ServerName = service
		qlog.Warnf("%s %s TLS mode, verify peer: %t, cipher plugin: %s.",
			protocol, service, sslConfig.VerifyPeer, sslConfig.CipherPlugin)
	}
	var command string
	if service != "" {
		command = strings.Join([]string{common.Consumer, service}, ".")
	}
	return f(Options{
		Service:   service,
		TLSConfig: tlsConfig,
		PoolSize:  GetMaxIdleCon(protocol),
		Failure:   GetFailureMap(protocol),
		Timeout:   config.GetTimeoutDurationFromArchaius(command, common.Consumer),
		Endpoint:  endpoint,
	})
}
func generateKey(protocol, service, endpoint string) string {
	return protocol + service + endpoint
}

// GetClient is to get the client based on protocol, service,endpoint name
func GetClient(protocol, service, endpoint string) (ProtocolClient, error) {
	var c ProtocolClient
	var err error
	key := generateKey(protocol, service, endpoint)
	sl.RLock()
	c, ok := clients[key]
	sl.RUnlock()
	if !ok {
		qlog.Info("Create client for " + protocol + ":" + service + ":" + endpoint)
		c, err = CreateClient(protocol, service, endpoint)
		if err != nil {
			return nil, err
		}
		sl.Lock()
		clients[key] = c
		sl.Unlock()
	}
	return c, nil
}

//Close close a client conn
func Close(protocol, service, endpoint string) error {
	key := generateKey(protocol, service, endpoint)
	sl.RLock()
	c, ok := clients[key]
	sl.RUnlock()
	if !ok {
		return ErrClientNotExist
	}
	if err := c.Close(); err != nil {
		qlog.Errorf("can not close client %s:%s%:s, err [%s]", protocol, service, endpoint, err.Error())
		return err
	}
	sl.Lock()
	delete(clients, key)
	sl.Unlock()
	return nil
}

// SetTimeoutToClientCache set timeout to client
func SetTimeoutToClientCache(spec *model.IsolationWrapper) {
	sl.Lock()
	defer sl.Unlock()
	for _, client := range clients {
		if client != nil {
			if v, ok := spec.Consumer.AnyService[client.GetOptions().Service]; ok {
				client.ReloadConfigs(Options{Timeout: time.Duration(v.TimeoutInMilliseconds) * time.Millisecond})
			} else {
				client.ReloadConfigs(Options{Timeout: time.Duration(spec.Consumer.TimeoutInMilliseconds) * time.Millisecond})
			}
		}
	}
}

// EqualOpts equal newOpts and oldOpts
func EqualOpts(oldOpts, newOpts Options) Options {
	if newOpts.Timeout != oldOpts.Timeout {
		oldOpts.Timeout = newOpts.Timeout
	}

	if newOpts.PoolSize != 0 {
		oldOpts.PoolSize = newOpts.PoolSize
	}
	if newOpts.PoolTTL != 0 {
		oldOpts.PoolTTL = newOpts.PoolTTL
	}
	if newOpts.TLSConfig != nil {
		oldOpts.TLSConfig = newOpts.TLSConfig
	}
	oldOpts.Failure = newOpts.Failure
	return oldOpts
}
