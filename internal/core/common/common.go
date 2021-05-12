package common

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/leon-yc/ggs/pkg/qlog"
)

// constant for provider and consumer
const (
	Provider = "Provider"
	Consumer = "Consumer"
)

const (
	// ScopeFull means service is able to access to another app's service
	ScopeFull = "full"
	// ScopeApp means service is not able to access to another app's service
	ScopeApp = "app"
)

// constant for micro service environment parameters
const (
	Env = "go-chassis_ENV"

	EnvNodeIP      = "HOSTING_SERVER_IP"
	EnvSchemaRoot  = "SCHEMA_ROOT"
	EnvProjectID   = "GGS_PROJECT_ID"
	EnvGGSEndpoint = "PAAS_GGS_ENDPOINT"
)

// constant environment keys service center, config center, monitor server addresses
const (
	CseRegistryAddress     = "GGS_REGISTRY_ADDR"
	CseConfigCenterAddress = "GGS_CONFIG_CENTER_ADDR"
	CseMonitorServer       = "GGS_MONITOR_SERVER_ADDR"
	EnvGGSSCEndpoint       = "PAAS_GGS_SC_ENDPOINT"
	EnvGGSCCEndpoint       = "PAAS_GGS_CC_ENDPOINT"
)

// env connect with "." like service_description.name and service_description.version which can not be used in k8s.
// So we can not use archaius to set env.
// To support this declaring constant for service name and version
// constant for service name and version.
const (
	ServiceName = "SERVICE_NAME"
	Version     = "VERSION"
)

// constant for microservice environment
const (
	EnvValueDev  = "development"
	EnvValueProd = "production"
)

// constant for secure socket layer parameters
const (
	SslCipherPluginKey = "cipherPlugin"
	SslVerifyPeerKey   = "verifyPeer"
	SslCipherSuitsKey  = "cipherSuits"
	SslProtocolKey     = "protocol"
	SslCaFileKey       = "caFile"
	SslCertFileKey     = "certFile"
	SslKeyFileKey      = "keyFile"
	SslCertPwdFileKey  = "certPwdFile"
	AKSKCustomCipher   = "ggs.credentials.akskCustomCipher"
)

// constant for protocol types
const (
	ProtocolRest    = "rest"
	ProtocolHighway = "highway"
	LBSessionID     = "go-chassisLB"
	ProtocolGrpc    = "grpc"
)

// configuration placeholders
const (
	PlaceholderInternalIP = "$INTERNAL_IP"
)

// SessionNameSpaceKey metadata session namespace key
const SessionNameSpaceKey = "_Session_Namespace"

// SessionNameSpaceDefaultValue default session namespace value
const SessionNameSpaceDefaultValue = "default"

// DefaultKey default key
const DefaultKey = "default"

// DefaultValue default value
const DefaultValue = "default"

// BuildinTagApp build tag for the application
const BuildinTagApp = "app"

// BuildinTagVersion build tag version
const BuildinTagVersion = "version"

// BuildinLabelVersion build label for version
const BuildinLabelVersion = BuildinTagVersion + ":" + LatestVersion

// CallerKey caller key
const CallerKey = "caller"

const (
	// HeaderSourceName is constant for header source name
	HeaderSourceName = "x-ggs-src-microservice"
	// HeaderXCseContent is constant for header , get some json msg about HeaderSourceName like {"k":"v"}
	HeaderXCseContent = "x-ggs-context"
	// HeaderSourceName is constant for header service name for sidecar
	HeaderXSidecar = "X-Ggs-Meshservice"
)

const (
	// SidecarAddress is constant for sidecar agent address.
	SidecarAddress = "127.0.0.1:8102"
)

// constant string for route type
const (
	RouteDiscovery = "discovery" //指定采用服务发现
	RouteDirect    = "direct"    //指定采用SLB或ipport
	RouteSidecar   = "sidecar"   //指定采用sidecar
	RouteDefault   = ""          //自动判断。(如果url.Host含有"."，则判为SLB/ipport，否则，判为服务发现)
)

// constant string for API path.
const (
	DefaultHealthzPath = "/ping"    //默认的健康检测接口
	DefaultMetricsPath = "/metrics" //默认的metrics接口
)

const (
	// RestMethod is the http method for restful protocol
	RestMethod = "method"
)

// constant for default application name and version
const (
	DefaultApp        = "default"
	DefaultVersion    = "v0.0.1"
	LatestVersion     = "latest"
	AllVersion        = "0+"
	DefaultStatus     = "UP"
	TESTINGStatus     = "TESTING"
	DefaultLevel      = "BACK"
	DefaultHBInterval = 30
)

//constant used
const (
	HTTP   = "http"
	HTTPS  = "https"
	JSON   = "application/json"
	Create = "CREATE"
	Update = "UPDATE"
	Delete = "DELETE"

	Client           = "client"
	File             = "File"
	DefaultTenant    = "default"
	DefaultChainName = "default"

	FileRegistry      = "File"
	DefaultUserName   = "default"
	DefaultDomainName = "default"
	DefaultProvider   = "default"

	TRUE  = "true"
	FALSE = "false"
)

// const default config for config-center
const (
	DefaultRefreshMode = 1
)

//ContextHeaderKey is the unified key of header value in context
//all protocol integrated with go chassis must set protocol header into context in this context key
type ContextHeaderKey struct{}

// NewContext transforms a metadata to context object
func NewContext(m map[string]string) context.Context {
	if m == nil {
		return context.WithValue(context.Background(), ContextHeaderKey{}, make(map[string]string, 0))
	}
	return context.WithValue(context.Background(), ContextHeaderKey{}, m)
}

// WithContext sets the KV and returns the context object
func WithContext(ctx context.Context, key, val string) context.Context {
	if ctx == nil {
		return context.WithValue(context.Background(), ContextHeaderKey{}, map[string]string{
			key: val,
		})
	}
	at, ok := ctx.Value(ContextHeaderKey{}).(map[string]string)
	if !ok {
		qlog.Trace("context header key does not has map, re-create new context")
		return context.WithValue(ctx, ContextHeaderKey{}, map[string]string{
			key: val,
		})
	}
	at[key] = val
	return ctx
}

// FromContext return the headers which should be send to provider
// through transport
func FromContext(ctx context.Context) map[string]string {
	if ctx == nil {
		return make(map[string]string, 0)
	}
	at, ok := ctx.Value(ContextHeaderKey{}).(map[string]string)
	if !ok {
		return make(map[string]string, 0)
	}
	return at
}

// GetXGGSContext  get x-ggs-context from req.header
func GetXGGSContext(k string, r *http.Request) string {
	if r == nil || r.Header == nil {
		qlog.Trace("get x-ggs-header failed , request(request.Header) is nil or  key is empty, please check its")
		return ""
	}
	cseContextStr := r.Header.Get(HeaderXCseContent)
	if cseContextStr == "" {
		return r.Header.Get(k)
	}

	var m map[string]string
	err := json.Unmarshal([]byte(cseContextStr), &m)
	if err != nil {
		qlog.Tracef("get x-ggs-header form req failed , error : %v", err)
		return ""
	}
	return m[k]
}

// SetXGGSContext  set value into x-ggs-context
func SetXGGSContext(vm map[string]string, r *http.Request) {
	if len(vm) <= 0 || vm == nil || r == nil {
		qlog.Trace("set x-ggs-header into req failed ,because one of key,value and request is empty(nil) or all empty(nil)")
		return
	}
	if r.Header == nil {
		r.Header = make(map[string][]string)
	}
	b, err := json.Marshal(vm)
	if err != nil {
		qlog.Tracef("set value to x-ggs-context failed , error : %v ", err.Error())
		return
	}
	r.Header.Set(HeaderXCseContent, string(b))
}
