package main

import (
	"fmt"

	// Check Patches pending
	centos "gitlab.esa.int/esait/sensu-check-critical-updates/centos"
	redhat "gitlab.esa.int/esait/sensu-check-critical-updates/redhat"
	ubuntu "gitlab.esa.int/esait/sensu-check-critical-updates/ubuntu"

	// Check how long the patches are pending
	pending "gitlab.esa.int/esait/sensu-check-critical-updates/pendingtime"

	// Normal inports
	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	"github.com/sensu/sensu-go/types"
	"gopkg.in/ini.v1"
)

// Config represents the check plugin config.
type Config struct {
	sensu.PluginConfig
	secCntWarn         int
	secCntCrit         int
	daysSincePatchWarn int
	daysSincePatchCrit int
	patchFilePath      string
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
			Path:      "sec-cnt-crit",
			Env:       "SEC_CNT_CRIT",
			Argument:  "sec-cnt-crit",
			Shorthand: "s",
			Default:   5,
			Usage:     "The number of security patches pending before reporting warning. (-1 to ignore)",
			Value:     &plugin.secCntCrit,
		},
		&sensu.PluginConfigOption{
			Path:      "sec-cnt-warn",
			Env:       "SEC_CNT_WARN",
			Argument:  "sec-cnt-warn",
			Shorthand: "w",
			Default:   10,
			Usage:     "The number of security patches pending before reporting warning. (-1 to ignore)",
			Value:     &plugin.secCntWarn,
		},
		&sensu.PluginConfigOption{
			Path:      "days-since-patch-warn",
			Env:       "DAYS-SINCE-PATCH-WARN",
			Argument:  "days-since-patch-warn",
			Shorthand: "p",
			Default:   30,
			Usage:     "Number of days patches are pending, before starting to report warning.",
			Value:     &plugin.daysSincePatchWarn,
		},
		&sensu.PluginConfigOption{
			Path:      "days-since-patch-crit",
			Env:       "DAYS-SINCE-PATCH-CRIT",
			Argument:  "days-since-patch-crit",
			Shorthand: "P",
			Default:   60,
			Usage:     "Number of days patches are pending, before starting to report critical.",
			Value:     &plugin.daysSincePatchCrit,
		},
		&sensu.PluginConfigOption{
			Path:      "patch-file-path",
			Env:       "patch-file-path",
			Argument:  "patch-file-path",
			Shorthand: "f",
			Default:   "/var/tmp/lastNoPending.run",
			Usage:     "Path to keep the file that is used to check the last time the amount of patches was 0.",
			Value:     &plugin.patchFilePath,
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

	osID := cfg.Section("").Key("ID").String()
	osLike := cfg.Section("").Key("ID_LIKE").String()
	//fmt.Println("OS-LIKE", osLike)

	return osID, osLike, nil
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
	osID, osRelease, err := getOSRelease()

	if err != nil {
		return 0, err
	}

	var sev int
	var checkErr error
	var numPatch int
	var numSec int
	var numCrit int
	var numImp int
	if osID == "ubuntu" || osID == "linuxmint" || osID == "debian" {
		sev, numPatch, numSec, numImp, numCrit, checkErr = ubuntu.CheckPatch(plugin.secCntWarn, plugin.secCntCrit)
	} else if osID == "rhel" {
		sev, numPatch, numSec, numImp, numCrit, checkErr = redhat.CheckPatch(plugin.secCntWarn, plugin.secCntCrit)

	} else if osID == "centos" {
		sev, numPatch, numSec, numImp, numCrit, checkErr = centos.CheckPatch(plugin.secCntWarn, plugin.secCntCrit)
	} else {
		return 0, fmt.Errorf("OS %s not supported", osRelease)
	}

	// Check if patches have been outstanding for too long
	patchStat, lastPatch, err := pending.PendingTime(plugin.patchFilePath, numPatch, plugin.daysSincePatchWarn, plugin.daysSincePatchCrit)
	if err != nil {
		fmt.Println("Pending file failed:", err)
		sev = sensu.CheckStateWarning
	}

	// If the outstanding patches criticality is higher than the security implications use the former
	if patchStat > sev {
		sev = patchStat
	}

	if sev == sensu.CheckStateOK {
		//File does not exist. OS is NOT indicating reboot required. Return OK
		fmt.Printf("%s OK: patches %d security %d important %d critical %d days %d.\n", plugin.PluginConfig.Name, numPatch, numSec, numImp, numCrit, lastPatch)
		return sensu.CheckStateOK, nil
	} else if sev == sensu.CheckStateWarning {
		fmt.Printf("%s WARNING: patches %d security %d important %d critical %d days %d.\n", plugin.PluginConfig.Name, numPatch, numSec, numImp, numCrit, lastPatch)
		return sensu.CheckStateWarning, nil
	} else if sev == sensu.CheckStateCritical {
		fmt.Printf("%s CRITICAL: patches %d security %d important %d critical %d days %d.\n", plugin.PluginConfig.Name, numPatch, numSec, numImp, numCrit, lastPatch)
		return sensu.CheckStateCritical, nil
	} else {
		//If the file exists, OS is indicating reboot required. Return Warning.
		//Maybe also return list of packages?

		fmt.Printf("%s UNKNOWN: %s %d\n", plugin.PluginConfig.Name, checkErr, sev)
		return sensu.CheckStateUnknown, nil
	}
}
