package fileutil

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/leon-yc/ggs/pkg/qlog"
)

const (
	//GgsConfDir is constant of type string
	GgsConfDir = "GGS_CONF_DIR"
	//GgsHome is constant of type string
	GgsHome = "GGS_HOME"
	//SchemaDirectory is constant of type string
	SchemaDirectory = "schema"
)

const (
	//App is a constant of type string
	App = "app.yaml"
	//Advance is a constant of type string
	Advanced = "advanced.yaml"

	//Global is a constant of type string
	Global = "chassis.yaml"
	//LoadBalancing is constant of type string
	LoadBalancing = "load_balancing.yaml"
	//RateLimiting is constant of type string
	RateLimiting = "rate_limiting.yaml"
	//Definition is constant of type string
	Definition = "microservice.yaml"
	//Hystric is constant of type string
	Hystric = "circuit_breaker.yaml"
	//PaasLager is constant of type string
	PaasLager = "log.yaml"
	//TLS is constant of type string
	TLS = "tls.yaml"
	//Monitoring is constant of type string
	Monitoring = "monitoring.yaml"
	//Auth is constant of type string
	Auth = "auth.yaml"
	//Tracing is constant of type string
	Tracing = "tracing.yaml"
	//Router is constant of type string
	Router = "router.yaml"
)

var configDir string
var homeDir string
var once sync.Once

//GetWorkDir is a function used to get the working directory
func GetWorkDir() (string, error) {
	wd, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", err
	}
	return wd, nil
}

func initDir() {
	if h := os.Getenv(GgsHome); h != "" {
		homeDir = h
	} else {
		//wd, err := GetWorkDir()
		//if err != nil {
		//	panic(err)
		//}
		//homeDir = wd
		homeDir = "."
	}

	// set conf dir, GGS_CONF_DIR has highest priority
	if confDir := os.Getenv(GgsConfDir); confDir != "" {
		configDir = confDir
	} else {
		// GGS_HOME has second most high priority
		configDir = filepath.Join(homeDir, "conf")
	}
	qlog.Infof("got conf dir: %s", configDir)
}

//GgsHomeDir is function used to get the home directory of ggs
func GgsHomeDir() string {
	once.Do(initDir)
	return homeDir
}

//GetConfDir is a function used to get the configuration directory
func GetConfDir() string {
	once.Do(initDir)
	return configDir
}

//SetConfDir is a function used to set the configuration directory
func SetConfDir(dir string) {
	configDir = dir
}

//CircuitBreakerConfigPath is a function used to join .yaml file name with configuration path
func CircuitBreakerConfigPath() string {
	return filepath.Join(GetConfDir(), Hystric)
}

//GetDefinition is a function used to join .yaml file name with configuration path
func GetDefinition() string {
	return filepath.Join(GetConfDir(), Definition)
}

//LoadBalancingConfigPath is a function used to join .yaml file name with configuration directory
func LoadBalancingConfigPath() string {
	return filepath.Join(GetConfDir(), LoadBalancing)
}

//RateLimitingFile is a function used to join .yaml file name with configuration directory
func RateLimitingFile() string {
	return filepath.Join(GetConfDir(), RateLimiting)
}

//TLSConfigPath is a function used to join .yaml file name with configuration directory
func TLSConfigPath() string {
	return filepath.Join(GetConfDir(), TLS)
}

//MonitoringConfigPath is a function used to join .yaml file name with configuration directory
func MonitoringConfigPath() string {
	return filepath.Join(GetConfDir(), Monitoring)
}

//MicroserviceDefinition is a function used to join .yaml file name with configuration directory
func MicroserviceDefinition(microserviceName string) string {
	return filepath.Join(GetConfDir(), microserviceName, Definition)
}

//MicroServiceConfigPath is a function used to join .yaml file name with configuration directory
func MicroServiceConfigPath() string {
	return filepath.Join(GetConfDir(), Definition)
}

//GlobalConfigPath is a function used to join .yaml file name with configuration directory
func GlobalConfigPath() string {
	return filepath.Join(GetConfDir(), Global)
}

//LogConfigPath is a function used to join .yaml file name with configuration directory
func LogConfigPath() string {
	return filepath.Join(GetConfDir(), PaasLager)
}

//RouterConfigPath is a function used to join .yaml file name with configuration directory
func RouterConfigPath() string {
	return filepath.Join(GetConfDir(), Router)
}

//AuthConfigPath is a function used to join .yaml file name with configuration directory
func AuthConfigPath() string {
	return filepath.Join(GetConfDir(), Auth)
}

//AppConfigPath is a function used to join .yaml file name with configuration directory
func AppConfigPath() string {
	return filepath.Join(GetConfDir(), App)
}

//AdvanceConfigPath is a function used to join .yaml file name with configuration directory
func AdvancedConfigPath() string {
	return filepath.Join(GetConfDir(), Advanced)
}

//TracingPath is a function used to join .yaml file name with configuration directory
func TracingPath() string {
	return filepath.Join(GetConfDir(), Tracing)
}

//SchemaDir is a function used to join .yaml file name with configuration path
func SchemaDir(microserviceName string) string {
	return filepath.Join(GetConfDir(), microserviceName, SchemaDirectory)
}
