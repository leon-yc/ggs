package runtime

import (
	"io/ioutil"
	"strings"

	"github.com/leon-yc/ggs/pkg/qlog"
)

//Status
const (
	StatusRunning = "UP"
	StatusDown    = "DOWN"
)

//HostName is the host name of service host
var HostName string

//ServiceID is the service id in registry service
var ServiceID string

//ServiceName represent self name
var ServiceName string

//Environment is usually represent as development, testing, production and  acceptance
var Environment string

//Schemas save schema file names(schema IDs)
var Schemas []string

//App is app info
var App string

//Version is version info
var Version string

//MD is service metadata
var MD map[string]string

//InstanceMD is instance metadata
var InstanceMD map[string]string

//InstanceID is the instance id in registry service
var InstanceID string

//InstanceStatus is the current status of instance
var InstanceStatus string

// InsideDocker is the flag if current process running in docker
var InsideDocker bool

//Init initialize runtime info
func Init() error {
	inDocker := hasCGroupDocker()
	if inDocker {
		qlog.Info("running in DOCKER")
	}
	InsideDocker = inDocker
	return nil
}

func hasCGroupDocker() bool {
	bytes, err := ioutil.ReadFile("/proc/self/cgroup")
	if err != nil {
		return false
	}
	cGroupContent := string(bytes)
	return strings.Contains(cGroupContent, "docker")
}
