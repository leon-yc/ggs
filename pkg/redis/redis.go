// pkg/redis库
// 功能简介:
// 1.开箱即用地拿到go-redis的Client，自动从配置文件解析配置初始化go-redis，使用方不需要再次读取配置，传递参数;
// 2.自动集成了trace;
//
// 使用示例:
// //...
// //程序初始化阶段，判断redis配置及网络连接的正确性.  [可选]
// if err := redis.CheckValid(); err!=nil {
//     panic(err)
// }
// //...
// redisName := xxx //配置中的redis名称
// ctx := xxx       //上下文传递的context.Context
// redisCli := redis.Client(redisName).WithContext(ctx)  //取到go-redis的Client对象(总是非nil)
// //...            //开始使用go-redis
//

package redis

import (
	"context"

	"github.com/leon-yc/ggs/pkg/errors"
	rds "github.com/go-redis/redis/v7"
)

type configWrap struct {
	Configs map[string]*Config `yaml:"redis"`
}

type Config struct {
	Addr         string `yaml:"addr"`
	Passwd       string `yaml:"password"`
	DB           int    `yaml:"database"`
	DialTimeout  int    `yaml:"dial_timeout"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
	PoolSize     int    `yaml:"pool_size"`
	MinIdleConns int    `yaml:"min_idle_conns"`
	MaxRetries   int    `yaml:"max_retries"`
}

// CheckValid checks redis clients' validity
// If no names are provided, all configed redises will be checked.
func CheckValid(names ...string) error {
	var redisNames []string
	if len(names) > 0 {
		redisNames = names
	} else {
		for k := range defaultMgr.getConfigs() {
			redisNames = append(redisNames, k)
		}
	}

	for _, name := range redisNames {
		if _, err := Client(name).Ping().Result(); err != nil {
			return errors.Wrapf(err, "redis(%s) not valid", name)
		}
	}
	return nil
}

// Client return a go-redis' Client connected to the redis-server of {name}
func Client(name string) *rds.Client {
	return getClient(name)
}

// ClientWithTrace return a go-redis' Client connected to the redis-server of {name} and with trace injected
func ClientWithTrace(ctx context.Context, name string) *rds.Client {
	return getClient(name).WithContext(ctx)
}

// GetConfig return the redis config of {name}
func GetConfig(name string) *Config {
	return defaultMgr.getConfigs()[name]
}

func getClient(name string) *rds.Client {
	cli := defaultMgr.obtainClient(name)
	if cli != nil {
		return cli
	}

	dummyCli := rds.NewClient(&rds.Options{Addr: "x0123456789"})
	return dummyCli
}
