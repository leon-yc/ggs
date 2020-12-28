package gorm

import (
	"context"
	"github.com/leon-gopher/qulibs/gorm_v2"
	"sync"
	"time"

	"github.com/leon-yc/ggs/pkg/conf"
	"github.com/leon-yc/ggs/pkg/qlog"
)

var (
	gormManager     *gorm.Manager
	gormManagerOnce sync.Once
)

type configWrap struct {
	Configs map[string]*Config `yaml:"gorm"`
}

type Config struct {
	Driver       string `yaml:"driver"`
	DSN          string `yaml:"dsn"`
	DialTimeout  int    `yaml:"dial_timeout"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
	MaxOpenConns int    `yaml:"max_open_conns"`
	MaxIdleConns int    `yaml:"max_idle_conns"`
	MaxLifeConns int    `yaml:"max_life_conns"`
	DebugSQL     bool   `yaml:"debug_sql"`
}

// Client  return a gorm' Client connected to the mysql-server of {name}
func Client(name string) *gorm.Client {
	return getClient(name)
}

// ClientWithTrace return a gorm' Client connected to the mysql-server of {name} and with trace injected
func ClientWithTrace(ctx context.Context, name string) *gorm.TraceClient {
	return getClient(name).Trace(ctx)
}

func getClient(name string) *gorm.Client {
	gormManagerOnce.Do(initManager)

	client, err := gormManager.NewClient(name)
	if err != nil {
		qlog.Errorf("get gorm client err=%s", err)
		return nil
	}
	return client
}

func initManager() {
	cfgWrap := new(configWrap)
	if err := conf.Unmarshal(cfgWrap); err != nil {
		qlog.Errorf("unmarshal gorm config err=%s", err)
	}

	mgrConf := make(map[string]*gorm.Config)
	for k, v := range cfgWrap.Configs {
		fillDefaultConfig(v)
		mgrConf[k] = &gorm.Config{
			Driver:       v.Driver,
			DSN:          v.DSN,
			DialTimeout:  time.Duration(v.DialTimeout),
			ReadTimeout:  time.Duration(v.ReadTimeout),
			WriteTimeout: time.Duration(v.WriteTimeout),
			MaxOpenConns: v.MaxOpenConns,
			MaxIdleConns: v.MaxIdleConns,
			MaxLifeConns: v.MaxLifeConns,
			DebugSQL:     v.DebugSQL,
		}
	}

	mgrConfParam := gorm.ManagerConfig(mgrConf)
	gormManager = gorm.NewManager(&mgrConfParam)
}

func fillDefaultConfig(conf *Config) {
	if conf.Driver == "" {
		conf.Driver = "mysql"
	}
	if conf.DialTimeout == 0 {
		conf.DialTimeout = 5000
	}
	if conf.ReadTimeout == 0 {
		conf.ReadTimeout = 5000
	}
	if conf.WriteTimeout == 0 {
		conf.WriteTimeout = 3000
	}
	if conf.MaxOpenConns == 0 {
		conf.MaxOpenConns = 256
	}
	if conf.MaxIdleConns == 0 {
		conf.MaxIdleConns = 10
	}
}
