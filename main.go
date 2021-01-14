package main

import (
	"fmt"

	redhat "gitlab.esa.int/esait/sensu-check-critical-updates/redhat"
	ubuntu "gitlab.esa.int/esait/sensu-check-critical-updates/ubuntu"

	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	"github.com/sensu/sensu-go/types"
	"gopkg.in/ini.v1"
)

// Config represents the check plugin config.
type Config struct {
	sensu.PluginConfig
	filePath string
}

var (
	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-check-critical-updates",
			Short:    "Check OS to see if specified file exists",
			Keyspace: "sensu.io/plugins/sensu-check-critical-updates/config",
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

// Retrieve the OS family of the current Linux distribution
// Returns: string: ID, string: ID_LIKE, error
func getOSRelease() (string, string, error) {
	cfg, err := ini.Load("/etc/os-release")
	if err != nil {
		return "", "", err
	}

	osId := cfg.Section("").Key("ID").String()
	osLike := cfg.Section("").Key("ID_LIKE").String()
	//fmt.Println("OS-LIKE", osLike)

	return osId, osLike, nil
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
	osId, osRelease, err := getOSRelease()

	if err != nil {
		return 0, err
	}

	var sev int
	var checkErr error
	var num_patch int
	var num_sec int
	var num_crit int
	if osRelease == "ubuntu" {
		sev, num_patch,num_sec,num_crit,checkErr = ubuntu.CheckPatch()
	} else if osId == "rhel" {
		sev, num_patch,num_sec,num_crit,checkErr = redhat.CheckPatch()
	} else {
		return 0, fmt.Errorf("OS %s not supported", osRelease)
	}

	if sev == sensu.CheckStateOK {
		//File does not exist. OS is NOT indicating reboot required. Return OK
		fmt.Printf("%s OK: patches %d security %d critical %d.\n", plugin.PluginConfig.Name, num_patch, num_sec, num_crit)
		return sensu.CheckStateOK, nil
	} else if sev == sensu.CheckStateWarning {
		fmt.Printf("%s WARNING: patches %d security %d critical %d.\n", plugin.PluginConfig.Name, num_patch, num_sec, num_crit)
		return sensu.CheckStateWarning, nil
	} else if sev > sensu.CheckStateCritical {
		fmt.Printf("%s CRITICAL: patches %d security %d critical %d.\n", plugin.PluginConfig.Name, num_patch, num_sec, num_crit)
		return sensu.CheckStateCritical, nil
	} else {
		//If the file exists, OS is indicating reboot required. Return Warning.
		//Maybe also return list of packages?

		fmt.Printf("%s UNKNOWN: %s\n", plugin.PluginConfig.Name, checkErr)
		return sensu.CheckStateUnknown, nil
	}
}
