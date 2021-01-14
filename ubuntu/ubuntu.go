package ubuntu

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/sensu-community/sensu-plugin-sdk/sensu"
)

// Prefered shell for command line execution
const ShellToUse = "bash"

// checkPatchUbuntu checks if a file exists (...and is not a directory)
// Borrowed from: https://golangcode.com/check-if-a-file-exists/
// int: severity, int: total updates, int: security updates, int: critical updates
func CheckPatch() (int, int, int, int, error) {
	// info, err := os.Stat(filename)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// Get the severities
	// _, err1 := mapSecUpdates()
	// if err1 != nil {
	// 	fmt.Println(fmt.Errorf("Problem with security update info: %s\n", err1))
	// }

	// app := "/bin/yum info-sec|grep  'Critical:'"
	app := "/usr/bin/apt list --upgradeable"
	cmd := exec.Command(ShellToUse, "-c", app)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if (err != nil) {
		log.Printf("error: %v\n", err)
		return sensu.CheckStateUnknown, 0, 0, 0, fmt.Errorf("Apt run failed: %s", err)
	}

	num_patch := strings.Count(stdout.String(), "upgradable from")
	num_sec := strings.Count(stdout.String(), "-security")
	num_crit := 0

	// Poor mans monitoring
	// This is because Ubuntu security reporting over patches is useless
	var retState int = sensu.CheckStateOK
	if num_sec > 10 {
		retState = sensu.CheckStateCritical
	} else if num_sec > 0 {
		retState = sensu.CheckStateWarning
	}

	return retState, num_patch, num_sec, num_crit, nil
}
