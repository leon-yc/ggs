//Package initiator init necessary module
// before every other package init functions
package initiator

import (
	"github.com/leon-yc/ggs/third_party/forked/jpillora/overseer"
	"io/ioutil"
	"os"

	"github.com/go-chassis/openlog"

	"github.com/leon-yc/ggs/internal/core/log"
	"github.com/leon-yc/ggs/internal/pkg/util/fileutil"
	"github.com/leon-yc/ggs/pkg/qlog"
	"gopkg.in/yaml.v2"
)

var (
	// loggerOptions has the configuration about logging
	loggerOptions *log.Options
)

func init() {
	disableOpenlogging()
	initConfDir()
	InitLogger()
}

// InitLogger initiate config file and openlogging before other modules
func InitLogger() {
	err := ParseLoggerConfig(fileutil.LogConfigPath())
	//initialize log in any case
	if err != nil {
		log.Init(&log.Options{
			LoggerLevel:   "INFO",
			RollingPolicy: "size",
		})
		if os.IsNotExist(err) {
			qlog.Infof("[%s] not exist", fileutil.LogConfigPath())
		} else {
			qlog.Fatal(err)
		}
	} else {
		log.Init(&log.Options{
			Output:              loggerOptions.Output,
			LoggerLevel:         loggerOptions.LoggerLevel,
			LoggerFile:          loggerOptions.LoggerFile,
			DisableReportCaller: loggerOptions.DisableReportCaller,
			LogFormatText:       loggerOptions.LogFormatText,
			RollingPolicy:       loggerOptions.RollingPolicy,
			LogRotateDate:       loggerOptions.LogRotateDate,
			LogRotateSize:       loggerOptions.LogRotateSize,
			LogBackupCount:      loggerOptions.LogBackupCount,
		})

	}
}

// ParseLoggerConfig unmarshals the logger configuration file(log.yaml)
func ParseLoggerConfig(file string) error {
	loggerOptions = &log.Options{}
	err := unmarshalYamlFile(file, loggerOptions)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return err
}

//ListenCfg just for parse listenAddress
type ListenCfg struct {
	Ggs struct {
		Protocols map[string]struct {
			Listen string `yaml:"listenAddress"`
		} `yaml:"protocols"`
	} `yaml:"ggs"`
}

// ParseListenAddresses unmarshals the listen configuration(app.yaml)
func ParseListenAddresses() ([]string, error) {
	file := fileutil.AppConfigPath()
	cfg := ListenCfg{}
	err := unmarshalYamlFile(file, &cfg)
	if err != nil {
		return nil, err
	}

	var addresses []string
	for _, v := range cfg.Ggs.Protocols {
		addresses = append(addresses, v.Listen)
	}
	return addresses, nil
}

func RestartConfig() (overseer.Config,error) {
	file := fileutil.AppConfigPath()
	cfg := overseer.Config{}
	err := unmarshalYamlFile(file, &cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, err
}

func unmarshalYamlFile(file string, target interface{}) error {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(content, target)
}

func disableOpenlogging() {
	openlog.SetLogger(dummpLogger{})
}
