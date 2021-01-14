package redhat

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"errors"

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

	// app := "/bin/yum info-sec|grep  'Critical:'"
	app := "yum --security updateinfo list"
	cmd := exec.Command(ShellToUse, "-c", app)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if (err != nil) {
		return sensu.CheckStateUnknown, 0, 0, 0, errors.New(fmt.Sprintf("Yum failed: %s", err))
	}

	//fmt.Printf("stdout\n%s\n", stdout.String())
	//fmt.Printf("stderr\n%s\n", stderr.String())

	num_patch := strings.Count(stdout.String(), "upgradable from")
	num_sec := strings.Count(stdout.String(), "RHSA-")
	num_crit := strings.Count(stdout.String(), "Critical/Sec.")

	var retState int = sensu.CheckStateOK
	if num_crit > 0 {
		retState = sensu.CheckStateCritical
	} else if num_sec > 0 {
		retState = sensu.CheckStateWarning
	}
	return retState, num_patch, num_sec, num_crit, nil
}
