package main

import (
	"fmt"
	// "log"
	"os"

	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	"github.com/sensu/sensu-go/types"
)

// Config represents the check plugin config.
type Config struct {
	sensu.PluginConfig
	filePath string
}

var (
	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-check-file-exists",
			Short:    "Check OS to see if specified file exists",
			Keyspace: "sensu.io/plugins/sensu-check-file-exists/config",
		},
	}

	options = []*sensu.PluginConfigOption{
		&sensu.PluginConfigOption{
			Path:      "file-path",
			Env:       "FILE_PATH",
			Argument:  "file-path",
			Shorthand: "f",
			Default:   "/var/run/reboot-required",
			Usage:     "location of file is required.",
			Value:     &plugin.filePath,
		},
	}
)

func main() {
	check := sensu.NewGoCheck(&plugin.PluginConfig, options, checkArgs, executeCheck, false)
	check.Execute()
}

func checkArgs(event *types.Event) (int, error) {
	// Eventually add code to validate path input... for now, trust yourself

	// Some ideas...
	// if len(plugin.rebootFile) == 0 {
	// 	return sensu.CheckStateWarning, fmt.Errorf("--reboot-file or REBOOT_FILE environment variable is required")
	// }

	//For now, just say all is OK so check will Execute()
	return sensu.CheckStateOK, nil
}

func executeCheck(event *types.Event) (int, error) {

	if fileExists(plugin.filePath) {
		//If the file exists, OS is indicating reboot required. Return Warning.
		//Maybe also return list of packages?

		fmt.Printf("%s WARNING: %v found.\n", plugin.PluginConfig.Name, plugin.filePath)
		return sensu.CheckStateWarning, nil
	} else {
		//File does not exist. OS is NOT indicating reboot required. Return OK
		fmt.Printf("%s OK: %v NOT found.\n", plugin.PluginConfig.Name, plugin.filePath)
		return sensu.CheckStateOK, nil
	}

}

// fileExists checks if a file exists (...and is not a directory)
// Borrowed from: https://golangcode.com/check-if-a-file-exists/
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
