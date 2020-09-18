package conf

import (
	"bytes"
	"io/ioutil"

	"github.com/leon-yc/ggs/internal/pkg/util/fileutil"
	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-archaius/cast"
	"gopkg.in/yaml.v2"
)

// GetConfDir return the config dir
func GetConfDir() string {
	return fileutil.GetConfDir()
}

// Get is for to get the value of configuration key
func Get(key string) interface{} {
	return archaius.Get(key)
}

//GetValue return interface
func GetValue(key string) cast.Value {
	return archaius.GetValue(key)
}

// Exist check the configuration key existence
func Exist(key string) bool {
	return archaius.Exist(key)
}

// Unmarshal unmarshals the config into a Struct. Make sure that the tags
// on the fields of the structure are properly set.
func Unmarshal(obj interface{}) error {
	content, err := ioutil.ReadFile(fileutil.AppConfigPath())
	if err != nil {
		return err
	}
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	return decoder.Decode(obj)
}

// GetBool is gives the key value in the form of bool
func GetBool(key string, defaultValue bool) bool {
	return archaius.GetBool(key, defaultValue)
}

// GetFloat64 gives the key value in the form of float64
func GetFloat64(key string, defaultValue float64) float64 {
	return archaius.GetFloat64(key, defaultValue)
}

// GetInt gives the key value in the form of GetInt
func GetInt(key string, defaultValue int) int {
	return archaius.GetInt(key, defaultValue)
}

// GetString gives the key value in the form of GetString
func GetString(key string, defaultValue string) string {
	return archaius.GetString(key, defaultValue)
}

// GetConfigs gives the information about all configurations
func GetConfigs() map[string]interface{} {
	return archaius.GetConfigs()
}
