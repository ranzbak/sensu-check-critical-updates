package centos

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/sensu-community/sensu-plugin-sdk/sensu"
)

// Prefered shell for command line execution
const ShellToUse = "bash"

// checkPatchUbuntu checks if a file exists (...and is not a directory)
// Borrowed from: https://golangcode.com/check-if-a-file-exists/
// int: severity, int: total updates, int: security updates, int: critical updates
func CheckPatch(secWarn int, secCrit int) (int, int, int, int, int, error) {
	// info, err := os.Stat(filename)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// app := "/bin/yum info-sec|grep  'Critical:'"
    app := "yum  list updates -C -q"
	cmd := exec.Command(ShellToUse, "-c", app)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if (err != nil) {
		return sensu.CheckStateUnknown, 0, 0, 0, 0, fmt.Errorf("Yum failed: %s", err)
	}

	num_patch := strings.Count(stdout.String(), "Updated Packages")
	num_sec := 0 // No security information in CentOS :-(
	num_crit := 0
	num_imp := 0

	var retState int = sensu.CheckStateOK
	if num_crit > secCrit && secCrit != -1 {
		retState = sensu.CheckStateCritical
	} else if num_sec > secWarn && secWarn != -1 {
		retState = sensu.CheckStateWarning
	}
	return retState, num_patch, num_sec, num_imp, num_crit, nil
}
