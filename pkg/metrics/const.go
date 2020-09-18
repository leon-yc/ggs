package metrics

const (
	ErrorNum      = "error_log_total"
	ErrorNumLevel = "level"
	ErrorNumHelp  = "Total number of error log"

	ReqQPS     = "rest_server_responses_total"
	ReqQPSHelp = "Total number of RESTful responses on server side."

	ReqDuration     = "rest_server_request_duration_seconds_bucket"
	ReqDurationHelp = "The RESTful request latencies in seconds on server side."

	GrpcReqQPS     = "grpc_server_responses_total"
	GrpcReqQPSHelp = "Total number of GRPC responses on server side."

	GrpcReqDuration     = "grpc_server_request_duration_seconds_bucket"
	GrpcReqDurationHelp = "The GRPC request latencies in seconds on server side."

	ClientReqQPS     = "rest_client_responses_total"
	ClientReqQPSHelp = "Total number of RESTful responses on client side."

	ClientReqDuration     = "rest_client_request_duration_seconds_bucket"
	ClientReqDurationHelp = "The RESTful request latencies in seconds on client side."

	ClientGrpcReqQPS     = "grpc_client_responses_total"
	ClientGrpcReqQPSHelp = "Total number of GRPC responses on client side."

	ClientGrpcReqDuration     = "grpc_client_request_duration_seconds_bucket"
	ClientGrpcReqDurationHelp = "The GRPC request latencies in seconds on client side."

	ReqProtocolLable = "protocol"
	RespUriLable     = "uri"
	RespCodeLable    = "status"
	RespHandlerLable = "handler"
	RemoteLable      = "remote"

	//qps, duration for redis
	RedisReqCount     = "redis_count"
	RedisReqCountHelp = "Total number of request redis."
	RedisRedisName    = "redis_name"
	RedisReqCMD       = "cmd"
	RedisRespStatus   = "status"

	RedisReqDurationSecond     = "redis_duration"
	RedisReqDurationSecondHelp = "Latency of redis duration in second."
)
