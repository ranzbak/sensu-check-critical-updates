package ubuntu

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"
)


// Prefered shell for command line execution
const ShellToUse = "bash"

// The number of to be installed security patches that is counted as a critical
// This because Ubuntu does not have a system that is usable for checking criticality
// Of patches on mass without very elaborate infrastructure
const numSecIsCrit = 10

// checkPatchUbuntu checks if a file exists (...and is not a directory)
// Borrowed from: https://golangcode.com/check-if-a-file-exists/
// bool: Patches applied, int: total updates, int: security updates, int: critical updates
func CheckPatch() (int, int, int, error) {
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
		return 0, 0, 0, errors.New(fmt.Sprintf("Apt run failed: %s", err))
	}

	num_patch := strings.Count(stdout.String(), "upgradable from")
	num_sec := strings.Count(stdout.String(), "-security")

	// Poor mans monitoring
	// This is because Ubuntu security reporting over patches is useless
	var num_crit = 0
	if num_sec > numSecIsCrit {
		num_crit = 1
	}

	return num_patch, num_sec, num_crit, nil
}
