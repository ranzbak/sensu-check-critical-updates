package ubuntu

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/sensu-community/sensu-plugin-sdk/sensu"
)

// ShellToUse Prefered shell for command line execution
const ShellToUse = "/bin/bash"

// CheckPatch checks if a file exists (...and is not a directory)
// Borrowed from: https://golangcode.com/check-if-a-file-exists/
// int: severity, int: total updates, int: security updates, int: critical updates
func CheckPatch(secWarn int, secCrit int) (int, int, int, int, int, error) {
	// info, err := os.Stat(filename)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// app := "/bin/yum info-sec|grep  'Critical:'"
	app := "/usr/bin/apt list --upgradeable"
	cmd := exec.Command(ShellToUse, "-c", app)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if (err != nil) {
		log.Printf("error: %v\n", err)
		return sensu.CheckStateUnknown, 0, 0, 0, 0, fmt.Errorf("Apt run failed: %s", err)
	}

	numPatch := strings.Count(stdout.String(), "upgradable from")
	numSec := strings.Count(stdout.String(), "-security")
	numCrit := 0
	numImp := 0

	// Poor mans monitoring
	// This is because Ubuntu security reporting over patches is useless
	var retState int = sensu.CheckStateOK
	if numSec > secWarn {
		retState = sensu.CheckStateWarning
	}

	return retState, numPatch, numSec, numImp, numCrit, nil
}
