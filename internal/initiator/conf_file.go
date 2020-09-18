package initiator

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/leon-yc/ggs/internal/pkg/util/fileutil"
	"github.com/leon-yc/ggs/pkg/qlog"
)

var (
	requiredFiles = []string{
		fileutil.App,
	}
	_ = flag.String("c", "./conf", "config path")
)

const (
	ConfigCenterDirPrefix = "/data/etc/cc"
)

func initConfDir() {
	confDir, err := parseC(os.Args[1:])
	if err == nil && confDir != "" {
		fmt.Printf("got config dir from args: %s", confDir)
		os.Setenv(fileutil.GgsConfDir, confDir)
	}
}

// 和配置中心路径规范保持一致
func NameNormalize(name string) string {
	return strings.ReplaceAll(name, "-", "_")
}

func guessConfDir() string {
	serviceName := NameNormalize(getValueFromArgs("service.name"))

	dirPaths := make([]string, 0, 2)
	if serviceName != "" {
		dirPaths = append(dirPaths, filepath.Join(ConfigCenterDirPrefix, serviceName))
	}
	dirPaths = append(dirPaths, ConfigCenterDirPrefix)

	for _, dirPath := range dirPaths {
		if isExistRequiredFile(dirPath) {
			return dirPath
		}
		//qlog.Infof("the folder[%s] does not contain the required files", dirPath)
	}
	return ""
}

func getValueFromArgs(key string) string {
	for i, value := range os.Args {
		if i == 0 {
			continue
		}
		idx := strings.Index(value, "=")
		if (value[0] == '-') && (idx >= 2) {
			if key == value[1:idx] {
				return value[idx+1:]
			}
		}
	}

	return ""
}

func isExistRequiredFile(dirPath string) bool {
	for _, fileName := range requiredFiles {
		filePath := filepath.Join(dirPath, fileName)
		_, err := os.Stat(filePath)
		if err == nil {
			return true
		}
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			qlog.Warnf("filePath[%s] stat error: %v", filePath, err)
		}
	}

	return false
}

func parseC(args []string) (string, error) {
	for {
		if len(args) == 0 {
			return "", nil
		}
		s := args[0]
		if len(s) < 2 || s[0] != '-' {
			return "", nil
		}
		numMinuses := 1
		if s[1] == '-' {
			numMinuses++
			if len(s) == 2 { // "--" terminates the flags
				args = args[1:]
				return "", nil
			}
		}
		name := s[numMinuses:]
		if len(name) == 0 || name[0] == '-' || name[0] == '=' {
			return "", fmt.Errorf("bad flag syntax: %s", s)
		}

		// it's a flag. does it have an argument?
		args = args[1:]
		hasValue := false
		value := ""
		for i := 1; i < len(name); i++ { // equals cannot be first
			if name[i] == '=' {
				value = name[i+1:]
				hasValue = true
				name = name[0:i]
				break
			}
		}

		// It must have a value, which might be the next argument.
		if !hasValue && len(args) > 0 {
			// value is the next arg
			hasValue = true
			value, args = args[0], args[1:]
		}
		if !hasValue {
			return "", fmt.Errorf("flag needs an argument: -%s", name)
		}

		if name == "c" {
			return value, nil
		}
	}
}
