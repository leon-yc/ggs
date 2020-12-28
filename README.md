# 简介
ggs(Go Gin Service)是一个微服务开发框架（库），目标是让业务方在打造微服务架构时，只需关心业务逻辑开发，自动具备服务治理能力，以提升业务开发效率。

# 为什么使用 ggs

- ggs封装了程序启动、退出、部署时的方方面面，让业务方不再被这些"琐事"所困扰
- ggs封装了服务治理的各环节，并且与公司组件无缝对接，将业务方的服务治理开发成本降到最低
- ggs封装了redis、mysql、nsq的调用姿势，为业务方提供开箱即用的能力
- ggs充分考虑上手成本，http的使用方式完全等同于gin，grpc的使用方式也做到极简


# 特性
 - **Gin无缝集成**: http的使用习惯上，与**gin完全保持一致**。
 - **多协议多端口**: 同时支持**http、grpc**，且支持同时监听多个端口。
 - **服务注册与发现**:  集成公司consul，**自动具备**服务注册、服务发现能力。
 - **配置中心**:  集成公司配置中心，**一键生成**框架定制模版，且自动生成合理参数配置，支持**热更新**，修改的配置无需重新发布自动reload生效。
 - **限流**:  **自动具备**限流能力，使用方只需配置参数即可，provider、consumer皆支持。
 - **调用链追踪**: 集成公司trace，同时对接收环节、远程调用环节做了深度封装，**自动具备**链路追踪能力。
 - **Metrics**:  **自动生成**metrics数据，且与promethues自动对接。
 - **负载均衡**:  作为consumer方调用远端时，支持多种负载均衡策略，支持重试，**按需配置即可**。
 - **隔离容错**:  通过超时、熔断策略在运行时保护你的分布式系统，免于错误雪崩，**按需配置即可**。
 - **sidecar**:  需要接入sidecar时，**指定参数即可**。
 - **log库**:  提供**开箱即用**的log库，支持stdout/file输出、自动切割、json格式、添加结构化字段、interface。
 - **redis/mysql**:  封装**开箱即用**的redis/mysql库，支持trace。
 - **部署**:  与PIEE的新CI&CD结合，**简单配置**即可实现部署，无缝支持k8s部署。
 - **优雅处理**:  支持**优雅退出**, 支持**平滑重启**，再也不用担心发布时的流量丢失或报错，再也不用担心并行发布时的负载均衡问题。


# Get Started
- 推荐Go版本: >= 1.13
## 一 快速搭建篇
### 1.0 使用ggs框架的最短路径？
- 安装ggs-gen
```bash
go get -u github.com/leon-yc/ggs/cmd/ggs-gen
```
备注: 如果下载很慢或者失败，可以使用文档最下方的代理策略。
- 生成新应用
```bash
ggs-gen new myapp
```

### 1.1 如何快速搭建http服务？
1.参考代码:
```go
package main
import (
    "github.com/leon-yc/ggs"
    "github.com/gin-gonic/gin"
    "net/http"
)
func main() {
    //Init读取配置，并初始化框架内部组件 
    if err := ggs.Init(); err != nil {
        panic(err)
    }	
    //获取*gin.Engine，按gin的方式设置路由
    r, err := ggs.Gin()
    if err != nil {
        panic(err)
    }
    r.GET("/ping", func(c *gin.Context) {
        c.String(http.StatusOK, "pong")
    })
    //启动服务 
    ggs.Run()
}
```
2.配置文件./conf/app.yaml:
```yaml
service:
  name: RestServer #[MUST]服务名，需要保障唯一，pre/prd环境必须采用ops项目id
  environment: dev  #[dev, qa, pre, prd]
ggs:
  protocols: #[rest, grpc, rest-xxx, grpc-xxx]
    rest: #监听某个端口，提供restful服务
      listenAddress: "0.0.0.0:8868"
```

### 1.2 如何快速搭建grpc服务？
1.定义proto
```bash
syntax = "proto3";
package proto;

service Greeter {
    rpc SayHello (HelloReq) returns (HelloResp);
}
message HelloReq {
    string name = 1;
}
message HelloResp {
    string message = 1;
}
```

2.protoc-gen-ggs插件生成*.pb.go
```bash
//安装插件
go get github.com/leon-yc/ggs/cmd/protoc-gen-ggs
//生成*.pb.go (请调整对应的路径)
protoc -I app/api/proto app/api/proto/*.proto  --ggs_out=plugins=grpc:app/api/proto/
```

3.参考代码:
```go
package main
import (
    "github.com/leon-yc/ggs"
    pb yyc
    "context"
)

type Service struct{}
func (s *Service) SayHello(ctx context.Context, in *pb.HelloReq) (*pb.HelloResp, error) {
    return &pb.HelloResp{ Message: "Hello " + in.Name,}, nil
}

func main() {
    //Init读取配置，并初始化框架内部组件 
    if err := ggs.Init(); err != nil {
        panic(err)
    }
    //注册grpc路由
    pb.RegisterGreeterServer(&Service{})
    //启动服务
    ggs.Run()
}

```
4.配置文件./conf/app.yaml:
```yaml
service:
  name: GrpcServer #[MUST]服务名，需要保障唯一，pre/prd环境必须采用ops项目id
  environment: dev  #[dev, qa, pre, prd]
ggs.protocols: #[rest, grpc, rest-xxx, grpc-xxx]
  grpc: #监听某个端口，提供restful服务
    listenAddress: "0.0.0.0:9000"
  rest: #提供rest服务, 例如: /metrics接口
    listenAddress: "0.0.0.0:9001"
```

### 1.3 如何快速实现http远程调用?
```go
package main
import (
    "context"
    "github.com/leon-yc/ggs"
    "github.com/leon-yc/ggs/invoke/rest"
)
func main() { 
    if err := ggs.Init(); err != nil {
        panic(err)
    }
    resp, err := rest.ContextGet(context.Background(), "http://RestServer/ping")
    //...
}
```

### 1.4 如何快速实现grpc远程调用？
```go
package main
import (
    "context"
    "github.com/leon-yc/ggs"
    pb "github.com/leon-yc/ggs/examples/grpc/server/app/api/proto"
)
func main() {
    if err := ggs.Init(); err != nil {
        panic(err)
    }	
    // declare client
    client := pb.NewGreeterClient("GrpcServer-grpc")
    // invoke SayHello
    resp, err := client.SayHello(context.Background(), &pb.HelloReq{
        Name: "world.",
    })
    //...
}
```

### 1.5 如何支持平滑重启？
```go
package main
import (
    "github.com/leon-yc/ggs"
)
func main() {
    //对入口函数做一次封装
    ggs.GraceFork(Main)
}

func Main() {
    //正常main代码
    //...
}
```

### 1.6 如何自定义输出metrics?
- [输出metrics的方式](https://github.com/leon-yc/ggs/blob/master/pkg/metrics/README.md)

### 1.7 如何加载自定义配置?
- [读取配置的方式](https://github.com/leon-yc/ggs/blob/master/examples/config/README.md)

### 1.8 如何升级ggs框架?
cd /path/to/your/project/，执行:
```go
go get -v github.com/leon-yc/ggs
```

## 二 服务治理篇

### 2.1 如何实现服务注册？
conf/advanced.yaml中配置:
```yaml
ggs.service:
  registry:
      disabled: false #是否禁用, [true, false], {default: false}
      address: http://10.0.1.101:8500 #[MUST]consul地址
```

### 2.2 如何实现服务发现?
conf/advanced.yaml中配置:
```yaml
ggs.service:
    registry:
      disabled: false #是否禁用, [true, false], {default: false}
      address: http://10.0.1.101:8500 #[MUST]consul地址
```

### 2.3 如何实现trace?
conf/advanced.yaml中配置:
```yaml
ggs.tracing:
  disabled: false #是否禁用, {default: false}
  settings:
    samplingRate: 1.0 #采样率, [0-1.0], {default: 1.0}
    traceFileName: /data/logs/trace/trace.log #trace文件绝对路径, {default: /data/logs/trace/trace.log}
```

### 2.4 如何实现metrics?
conf/advanced.yaml中配置:
```yaml
ggs.metrics:
    enabled: true #是否开启, {default: false}
    apiPath: /metrics   #输出metrics的api
```

### 2.5 如何实现限流？
conf/advanced.yaml中配置:
```yaml
ggs.flowcontrol:
    Provider: #接收的流量限制
      qps:
        enabled: true #是否开启, {default: false}
        global.limit: 100000 #provider的全局限流qps, {default: 2147483647}
    Consumer: #远程调用的流量限制
      qps:
        enabled: true #是否开启, {default: false}, qps: {default: 2147483647}
        limit.FullServer: 10000                 #某服务
        limit.FullServer.rest./sayhi: 1000      #某服务的某API (优先级高)
```

### 2.6 如何实现重试?
conf/advanced.yaml中配置:
```yaml
ggs.loadbalance:
    strategy.name: RoundRobin #负载均衡策略，[RoundRobin,Random,WeightedResponse,SessionStickiness],{default:RoundRobin}
    retryEnabled: true #是否开启重试, {default: false}
    retryOnNext: 1 #"下一个"目标节点的重试最大次数, {default: 0}
    retryOnSame: 0 #同一个目标节点的重试最大次数 (总次数是: (retryOnSame+1)*(retryOnNext+1)), {default: 0}
```

### 2.7 如何实现超时?
conf/advanced.yaml中配置:
```yaml
ggs.isolation.Consumer: #异常隔离
  timeoutInMilliseconds: 1000 #远程调用的超时时间,单位:ms, {default: 1000}
  maxConcurrentRequests: 5000 #远程调用的最大并发数(goroutine), {default: 5000}
```

### 2.8 如何实现熔断?
conf/advanced.yaml中配置:
```yaml
ggs.circuitBreaker: #熔断
  scope: api #熔断范围, [api, instance, instance-api], {default: api}
  Consumer:
    enabled: true #是否开启熔断, {default: false}
    sleepWindowInMilliseconds: 15000 #熔断窗口期(熔断后到尝试恢复到间隔时间), 单位:ms, {default: 15000}
    requestVolumeThreshold: 50 #在一个窗口期内，是否check熔断的门槛次数，如果请求次数小于此值，即使100%失败，也不会触发熔断, {default: 50}
    errorThresholdPercentage: 50 #触发熔断的错误率, 单位:%, {default: 50}
```

### 2.9 如何通过sidecar访问远端?
```go
import "github.com/leon-yc/ggs/invoke"

//比普通的rest访问，多指定一个参数即可。
resp, err := rest.ContextGet(context.Background(), url, invoke.WithSidecar())
//...
```

## 三 公共服务调用篇

### 3.1 如何调用redis?
- 参见[pkg/redis](https://github.com/leon-yc/ggs/tree/master/pkg/redis)

### 3.2 如何调用mysql?
```go
import (
    yyc
    gogorm "github.com/jinzhu/gorm"
    //...
)

//...
//开箱即用地获取类gorm' Client
db := store.GormWithTrace(c.Request.Context(), "local")
//正常使用gorm
//...
```


# 四 参数配置
## 4.1 命令行参数
- -c=/path/to/conf/dir/ 指定配置文件目录。如果没有指定，则自动读取当前路径下的./conf。

## 4.2 配置文件路径
- dev: git维护, 当前conf/目录
- qa/pre/prd: 配置中心, /data/etc/cc/{ServiceName}/

## 4.3 完整的配置文件参数
- [docs/std-confs/app.yaml](https://github.com/leon-yc/ggs/tree/master/docs/std-confs/app.yaml)
  
  [MUST]服务基本信息、业务私有配置
- [docs/std-confs/advanced.yaml](https://github.com/leon-yc/ggs/tree/master/docs/std-confs/advanced.yaml)
  
  [OPTIONAL]服务治理相关的配置
- [docs/std-confs/log.yaml](https://github.com/leon-yc/ggs/tree/master/docs/std-confs/log.yaml)
  
  [OPTIONAL]log相关的配置
  
# 五 部署
##  CICD需要的参数
- CI构建命令（编译命令）
  - go mod模式（推荐）：
  ```bash
  make artifacts_mod
  ```
  - 非go mod模式：
  ```bash
  make artifacts
  ```
- CD运行命令
  - ECS发布
    - 优雅重启（推荐）（代码需要支持）:
    ```bash
    sh run.sh reload 
    ```
    - 普通重启:
    ```bash
    sh run.sh restart 
    ```
  - 容器发布
    ```bash
    #paas-project-id代表你的paas项目id
    #注意：路径中需要把"-"换成"_"
    ./paas-project-id -c /data/etc/cc/paas_project_id
    示例:
    ./techcenter-arch-ggs-k8s-demo -c /data/etc/cc/techcenter_arch_ggs_k8s_demo
    ```
- 健康检查路径
  /ping
- 监控服务路径
  /metrics
- 勾选: 配置中心与trace

# 参考示例
- [ggs-rest-example](https://github.com/leon-yc/ggs/ggs-rest-example)
- [examples/rest](https://github.com/leon-yc/ggs/tree/master/examples/rest)
- [examples/grpc](https://github.com/leon-yc/ggs/tree/master/examples/grpc)
- [examples/all](https://github.com/leon-yc/ggs/tree/master/examples/all)

# 参考源码
- [go-chassis](https://github.com/go-chassis/go-chassis)
- [gin](https://github.com/gin-gonic/gin)
- [grpc](https://grpc.io/)

# FAQ
- 依赖包下载慢或timeout?

  可以使用如下方式[代理](https://goproxy.cn/):
```bash
#If your Go version >= 1.13
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOPRIVATE=*.leon.net
#go on ...
```

- ggs框架交付给使用方的都有哪些东西？

  为了尽量解耦框架与业务，减少对业务开发的侵入，ggs没有提供很重的脚手架，以让模块边界明明白白，ggs-gen生成的项目，交付到使用方就以下几项:
  - ggs库(github.com/leon-yc/ggs)
  - 编译用的Makefile
  - 运行命令脚本(scripts/run.sh)
  - 配置文件(conf/)

- 我应该使用哪个ggs版本，升级是会有兼容问题吗？

  尽量使用最新版本。ggs升级时，可能会有个别的接口变动导致的编译问题，但一般只需稍微调整即可，如碰到任何兼容性问题，请联系我们，我们全人力负责纠正。



