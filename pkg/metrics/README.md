# metrics

- 开箱即用的metrics库
- 支持使用方输出自己的metrics数据
- metrics数据自动收集到公司到prometheus，无需额外运维工单

## 默认输出
目前框架内默认输出的数据包括:
- go的runtime数据
- provider的qps、状态码、响应时长
- consumer的qps、状态码、响应时长
- redis访问的qps、返回状态、响应时长
- error log的频率

具体字段详见const.go

## 示例
### go runtime
```bash
# HELP go_goroutines Number of goroutines that currently exist.
# TYPE go_goroutines gauge
go_goroutines 32
# HELP go_info Information about the Go environment.
# TYPE go_info gauge
go_info{version="go1.13"} 1
```
### provider qps
```bash
# HELP rest_server_responses_total Total number of RESTful responses on server side.
# TYPE rest_server_responses_total counter
rest_server_responses_total{protocol="rest",status="200",uri="/metrics"} 3
```

### provider response time
```bash
# HELP rest_server_request_duration_seconds_bucket The RESTful request latencies in seconds on server side.
# TYPE rest_server_request_duration_seconds_bucket histogram
rest_server_request_duration_seconds_bucket_bucket{protocol="rest",status="200",uri="/metrics",le="0.0005"} 1
rest_server_request_duration_seconds_bucket_bucket{protocol="rest",status="200",uri="/metrics",le="0.001"} 3
rest_server_request_duration_seconds_bucket_bucket{protocol="rest",status="200",uri="/metrics",le="0.005"} 3
rest_server_request_duration_seconds_bucket_bucket{protocol="rest",status="200",uri="/metrics",le="0.01"} 3
rest_server_request_duration_seconds_bucket_bucket{protocol="rest",status="200",uri="/metrics",le="0.025"} 3
rest_server_request_duration_seconds_bucket_bucket{protocol="rest",status="200",uri="/metrics",le="0.05"} 3
rest_server_request_duration_seconds_bucket_bucket{protocol="rest",status="200",uri="/metrics",le="0.1"} 3
rest_server_request_duration_seconds_bucket_bucket{protocol="rest",status="200",uri="/metrics",le="0.25"} 3
rest_server_request_duration_seconds_bucket_bucket{protocol="rest",status="200",uri="/metrics",le="0.5"} 3
rest_server_request_duration_seconds_bucket_bucket{protocol="rest",status="200",uri="/metrics",le="1"} 3
rest_server_request_duration_seconds_bucket_bucket{protocol="rest",status="200",uri="/metrics",le="2.5"} 3
rest_server_request_duration_seconds_bucket_bucket{protocol="rest",status="200",uri="/metrics",le="5"} 3
rest_server_request_duration_seconds_bucket_bucket{protocol="rest",status="200",uri="/metrics",le="10"} 3
rest_server_request_duration_seconds_bucket_bucket{protocol="rest",status="200",uri="/metrics",le="+Inf"} 3
rest_server_request_duration_seconds_bucket_sum{protocol="rest",status="200",uri="/metrics"} 0.001789716
rest_server_request_duration_seconds_bucket_count{protocol="rest",status="200",uri="/metrics"} 3
```

### consumer qps
```bash
# HELP rest_client_responses_total Total number of RESTful responses on client side.
# TYPE rest_client_responses_total counter
rest_client_responses_total{protocol="rest",remote="127.0.0.1:9000",status="0",uri="/sayhi"} 2
rest_client_responses_total{protocol="rest",remote="FullServer",status="200",uri="/sayhi"} 2
rest_client_responses_total{protocol="rest",remote="FullServer-rest-admin",status="200",uri="/admin"} 2
```

### consumer response time
```bash
# HELP rest_client_request_duration_seconds_bucket The RESTful request latencies in seconds on client side.
# TYPE rest_client_request_duration_seconds_bucket histogram
rest_client_request_duration_seconds_bucket_bucket{protocol="rest",remote="127.0.0.1:9000",status="0",uri="/sayhi",le="0.0005"} 1
rest_client_request_duration_seconds_bucket_bucket{protocol="rest",remote="127.0.0.1:9000",status="0",uri="/sayhi",le="0.001"} 2
rest_client_request_duration_seconds_bucket_bucket{protocol="rest",remote="127.0.0.1:9000",status="0",uri="/sayhi",le="0.005"} 2
rest_client_request_duration_seconds_bucket_bucket{protocol="rest",remote="127.0.0.1:9000",status="0",uri="/sayhi",le="0.01"} 2
rest_client_request_duration_seconds_bucket_bucket{protocol="rest",remote="127.0.0.1:9000",status="0",uri="/sayhi",le="0.025"} 2
rest_client_request_duration_seconds_bucket_bucket{protocol="rest",remote="127.0.0.1:9000",status="0",uri="/sayhi",le="0.05"} 2
rest_client_request_duration_seconds_bucket_bucket{protocol="rest",remote="127.0.0.1:9000",status="0",uri="/sayhi",le="0.1"} 2
rest_client_request_duration_seconds_bucket_bucket{protocol="rest",remote="127.0.0.1:9000",status="0",uri="/sayhi",le="0.25"} 2
rest_client_request_duration_seconds_bucket_bucket{protocol="rest",remote="127.0.0.1:9000",status="0",uri="/sayhi",le="0.5"} 2
rest_client_request_duration_seconds_bucket_bucket{protocol="rest",remote="127.0.0.1:9000",status="0",uri="/sayhi",le="1"} 2
rest_client_request_duration_seconds_bucket_bucket{protocol="rest",remote="127.0.0.1:9000",status="0",uri="/sayhi",le="2.5"} 2
rest_client_request_duration_seconds_bucket_bucket{protocol="rest",remote="127.0.0.1:9000",status="0",uri="/sayhi",le="5"} 2
rest_client_request_duration_seconds_bucket_bucket{protocol="rest",remote="127.0.0.1:9000",status="0",uri="/sayhi",le="10"} 2
rest_client_request_duration_seconds_bucket_bucket{protocol="rest",remote="127.0.0.1:9000",status="0",uri="/sayhi",le="+Inf"} 2
rest_client_request_duration_seconds_bucket_sum{protocol="rest",remote="127.0.0.1:9000",status="0",uri="/sayhi"} 0.0013316410000000002
rest_client_request_duration_seconds_bucket_count{protocol="rest",remote="127.0.0.1:9000",status="0",uri="/sayhi"} 2
```

### redis qps
```bash
# HELP redis_count Total number of request redis.
# TYPE redis_count counter
redis_count{cmd="get",redis_name="local",status="200"} 2
```

### redis response time
```bash
# HELP redis_duration Latency of redis duration in second.
# TYPE redis_duration histogram
redis_duration_bucket{cmd="get",redis_name="local",status="200",le="0.0005"} 0
redis_duration_bucket{cmd="get",redis_name="local",status="200",le="0.001"} 0
redis_duration_bucket{cmd="get",redis_name="local",status="200",le="0.005"} 0
redis_duration_bucket{cmd="get",redis_name="local",status="200",le="0.01"} 0
redis_duration_bucket{cmd="get",redis_name="local",status="200",le="0.025"} 1
redis_duration_bucket{cmd="get",redis_name="local",status="200",le="0.05"} 2
redis_duration_bucket{cmd="get",redis_name="local",status="200",le="0.1"} 2
redis_duration_bucket{cmd="get",redis_name="local",status="200",le="0.25"} 2
redis_duration_bucket{cmd="get",redis_name="local",status="200",le="0.5"} 2
redis_duration_bucket{cmd="get",redis_name="local",status="200",le="1"} 2
redis_duration_bucket{cmd="get",redis_name="local",status="200",le="2.5"} 2
redis_duration_bucket{cmd="get",redis_name="local",status="200",le="5"} 2
redis_duration_bucket{cmd="get",redis_name="local",status="200",le="10"} 2
redis_duration_bucket{cmd="get",redis_name="local",status="200",le="+Inf"} 2
redis_duration_sum{cmd="get",redis_name="local",status="200"} 0.049549481
redis_duration_count{cmd="get",redis_name="local",status="200"} 2
```

### error log count
```bash
# HELP error_log_total Total number of error log
# TYPE error_log_total counter
error_log_total{level="ERROR"} 2
```