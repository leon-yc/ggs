package redis

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-chassis/go-archaius"

	"github.com/leon-yc/ggs/pkg/conf"
	"github.com/leon-yc/ggs/pkg/qlog"
	rds "github.com/go-redis/redis/v7"
)

var defaultMgr redisMgr = redisMgr{
	rdsMap:  make(map[string]*rds.Client),
	configs: make(map[string]*Config),
}

type redisMgr struct {
	clientsMutex sync.Mutex
	rdsMap       map[string]*rds.Client
	configs      map[string]*Config
	configsOnce  sync.Once
	metricsOnce  sync.Once
}

func (mgr *redisMgr) obtainClient(name string) *rds.Client {
	cli := mgr.getClient(name)
	if cli != nil {
		return cli
	}
	return mgr.addClient(name)
}

func (mgr *redisMgr) getClient(name string) *rds.Client {
	mgr.clientsMutex.Lock()
	defer mgr.clientsMutex.Unlock()
	return mgr.rdsMap[name]
}

func (mgr *redisMgr) addClient(name string) *rds.Client {
	configs := mgr.getConfigs()
	cfg := configs[name]
	cli := rds.NewClient(&rds.Options{
		Network:      "tcp",
		Addr:         cfg.Addr,
		Password:     cfg.Passwd,
		DB:           cfg.DB,
		DialTimeout:  time.Duration(cfg.DialTimeout) * time.Millisecond,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Millisecond,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
	})
	cli.AddHook(newHook(name, cfg, archaius.GetBool("ggs.metrcs.disableRedisMetrics", false)))

	mgr.clientsMutex.Lock()
	mgr.rdsMap[name] = cli
	mgr.clientsMutex.Unlock()
	return cli
}

func (mgr *redisMgr) getConfigs() map[string]*Config {
	mgr.loadConfigs()
	return mgr.configs
}

func (mgr *redisMgr) loadConfigs() {
	mgr.configsOnce.Do(func() {
		cfgWrap := new(configWrap)
		if err := conf.Unmarshal(cfgWrap); err != nil {
			qlog.Error("unmarshal redis configs failed: ", err)
			return
		}

		for k, v := range cfgWrap.Configs {
			mgr.fillDefaultConfig(k, v)
		}
		mgr.configs = cfgWrap.Configs
	})
}

func (mgr *redisMgr) fillDefaultConfig(name string, c *Config) {
	if c.DialTimeout == 0 {
		c.DialTimeout = 5000
	}
	if c.ReadTimeout == 0 {
		c.ReadTimeout = 1000
	}
	if c.WriteTimeout == 0 {
		c.WriteTimeout = 1000
	}
	if c.MaxRetries == 0 {
		key := fmt.Sprintf("redis.%s.max_retries", name)
		if !conf.Exist(key) {
			c.MaxRetries = 1
		}
	}
}
