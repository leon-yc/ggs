## v1.1.11
- 支持provider端的api级别限流
- 支持通过http方式来触发reload

## v1.1.10
- 规范限流、熔断触发时的状态码

## v1.1.9
- 配置文件中支持全局开关sidecar
- make run增加proxy代理

## v1.1.8
- 支持访问远端的metrics数据
- 支持业务方输出自定义的metrics数据
- 支持输出访问redis的metrics数据
- ggs.Init()可以指定conf目录（单元测试可能用到）
- 简化CI构建命令

## v1.1.7
- 支持自动收集metrics数据到promethues，无需额外提工单；
- 限流过滤/ping,/metrics;
- 优化命令行参数；
- 提供是否在容器里的运行时判断；
- 优化ping，log；

## v1.1.6
- 支持部分远端请求走sidecar，部分不走；
- 优化CD运行脚本；

## v1.1.5
- fix脚手架的bug。

## v1.1.4
- 优化脚手架并补充文档。
- 增加获取配置数据的示例。
- 增加健康检测的ping接口。

## v1.1.3
- 增加生成基于ggs的脚手架。
- 增加pkg/qenv获取当前环境信息。

## v1.1.2
- 优化redis库的使用方式。

## v1.1.1
- http/grpc皆支持平滑重启。
- 修正默认超时时间。

## v1.1.0
- 明确外部可访问模块与ggs内部模块。
- 简化http远程调用方式。

## v1.0.7
- 支持重试条件的自定义配置。
- metrics输出使用方的error-log数目。

## v1.0.6
- 提供errors库，支持error wrap及error输出堆栈信息。
- 提供获取配置文件路径的接口。

## v1.0.5
- 限流实现由阻塞式改为否决式
- 支持k8s部署

## v1.0.4
- fix provider限流时，返回错误码不准确的问题
- 增加grpc的recover处理

## v1.0.3
- 与新CICD结合部署更顺畅
- 支持开启access.log

## v1.0.2
- 解决下载依赖包慢或timeout的问题；
- 支持http服务的平滑重启；

## v1.0.1
- 统一采用-c=/path/to/conf/来指定配置目录；
- 新增用于新CD的make命令，增加run.sh脚本；

## v1.0.0
- 无需额外代码，即可具备全套微服务治理能力；
- 支持快速搭建服务框架；
- 完全支持gin实现http服务；
- 支持grpc服务与调用，方式简洁；
- 开箱即用的log、redis、gorm，且自动实现了trace链；