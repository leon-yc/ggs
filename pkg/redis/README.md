# pkg/redis

## 功能简介
- 开箱即用地拿到go-redis的Client，自动从配置文件解析配置初始化go-redis，使用方不需要再次读取配置，传递参数;
- 自动集成了trace;

## 使用示例:
```go
import(
    ggsredis yyc
    "github.com/go-redis/redis/v7"
)

//...
//程序初始化阶段，判断redis配置及网络连接的正确性.  [可选]
if err := ggsredis.CheckValid(); err!=nil {
    panic(err)
}

//...

redisName := "local" //配置中的redis名称
ctx := xxx           //上下文传递的context.Context
redisCli := ggsredis.RedisWithTrace(ctx, redisName)          //【方式一】取到go-redis的Client对象(总是非nil)
//redisCli := ggsredis.Client(redisName).WithContext(ctx)    //【方式二】取到go-redis的Client对象(总是非nil)
//...            //开始使用go-redis
```
