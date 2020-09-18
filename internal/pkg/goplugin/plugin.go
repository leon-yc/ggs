package goplugin

import (
	"os"

	"github.com/leon-yc/ggs/internal/pkg/util/fileutil"
)

// LookupPlugin lookup plugin
// Caller needs to determine itself whether the plugin file exists
func LookupPlugin(name string) (string, error) {
	var pluginPath string
	var err error
	// firstly search plugin in {GgsHome}/lib
	pluginPath = fileutil.GgsHomeDir() + "/lib/" + name
	if _, err = os.Stat(pluginPath); err == nil {
		return pluginPath, nil
	}
	if !os.IsNotExist(err) {
		return "", err
	}

	// secondly search plugin in /usr/lib
	pluginPath = "/usr/lib/" + name
	if _, err = os.Stat(pluginPath); err == nil {
		return pluginPath, nil
	}
	return "", err
}

// LookUpSymbolFromPlugin looks up symbol from the plugin
func LookUpSymbolFromPlugin(plugName, symName string) (interface{}, error) {
	return lookUp(plugName, symName)
}
