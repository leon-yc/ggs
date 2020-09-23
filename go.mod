module github.com/leon-yc/ggs

require (
	github.com/AlecAivazis/survey/v2 v2.0.4
	github.com/aws/aws-sdk-go v1.25.42
	github.com/cenkalti/backoff v2.0.0+incompatible
	github.com/dolab/colorize v0.0.0-20180106055552-10753a0b4d68
	github.com/dolab/logger v0.0.0-20181130034249-dcb994406102
	github.com/gin-gonic/gin v1.6.3
	github.com/go-chassis/foundation v0.1.1-0.20200825060850-b16bf420f7b3
	github.com/go-chassis/go-archaius v0.24.0
	github.com/go-chassis/go-chassis v1.7.6 // indirect
	github.com/go-chassis/go-chassis-config v0.15.0
	github.com/go-chassis/go-restful-swagger20 v1.0.2-0.20191029071646-8c0119f661c5
	github.com/go-mesh/openlogging v1.0.1
	github.com/go-redis/redis/v7 v7.0.0-beta.4
	github.com/golang/protobuf v1.3.3
	github.com/golib/cli v1.3.1
	github.com/gorilla/websocket v1.4.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0
	github.com/hashicorp/go-version v1.0.0
	github.com/jinzhu/gorm v1.9.16 // indirect
	github.com/jpillora/opts v1.1.2
	github.com/jpillora/overseer v0.0.0-20190427034852-ce9055846616
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/nsqio/go-nsq v1.0.7
	github.com/opentracing/opentracing-go v1.1.0
	github.com/openzipkin/zipkin-go v0.2.2 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.1.0
	github.com/prometheus/common v0.6.0
	github.com/sirupsen/logrus v1.4.2
	github.com/smartystreets/goconvey v1.6.4
	github.com/spf13/viper v1.6.1 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/uber/jaeger-client-go v2.20.1+incompatible // indirect
	github.com/uber/jaeger-lib v2.2.0+incompatible // indirect
	go.uber.org/atomic v1.5.0 // indirect
	go.uber.org/automaxprocs v1.2.0
	go.uber.org/ratelimit v0.1.0
	golang.org/x/time v0.0.0-20190921001708-c4c64cad1fd0
	google.golang.org/grpc v1.24.0
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v8 v8.18.2 // indirect
	gopkg.in/yaml.v2 v2.2.8
	github.com/leon-gopher/discovery v1.0.1
	github.com/leon-gopher/gtracing v1.0.1
	github.com/leon-gopher/qulibs v1.0.1
)

replace (
	golang.org/x/crypto v0.0.0-20180820150726-614d502a4dac => github.com/golang/crypto v0.0.0-20180820150726-614d502a4dac
	golang.org/x/net v0.0.0-20180824152047-4bcd98cce591 => github.com/golang/net v0.0.0-20180824152047-4bcd98cce591
	golang.org/x/sys v0.0.0-20180824143301-4910a1d54f87 => github.com/golang/sys v0.0.0-20180824143301-4910a1d54f87
	golang.org/x/text v0.3.0 => github.com/golang/text v0.3.0
	golang.org/x/time v0.0.0-20180412165947-fbb02b2291d2 => github.com/golang/time v0.0.0-20180412165947-fbb02b2291d2
	google.golang.org/genproto v0.0.0-20180817151627-c66870c02cf8 => github.com/google/go-genproto v0.0.0-20180817151627-c66870c02cf8
)

go 1.14
