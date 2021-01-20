package redhat

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"

	"github.com/sensu-community/sensu-plugin-sdk/sensu"
)

// ShellToUse Prefered shell for command line execution
const ShellToUse = "bash"

func regexToNum(input string, reg string) (int, error) {
	compRegex := regexp.MustCompile(reg)
	match := compRegex.FindAllStringSubmatch(input, -1)

	if len(match) == 0 {
		return 0, nil
	}

	numStr := match[0][1]
	retval, err := strconv.Atoi(numStr)
	if err != nil {
		fmt.Println("No number found for", input, ":", err)
		return 0, err
	}

	return retval, nil
}

// CheckPatch checks if a file exists (...and is not a directory)
// Borrowed from: https://golangcode.com/check-if-a-file-exists/
// int: severity, int: total updates, int: security updates, int: critical updates
func CheckPatch(secWarn int, secCrit int) (int, int, int, int, int, error) {
	// info, err := os.Stat(filename)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// app := "/bin/yum info-sec|grep  'Critical:'"
	app := "yum updateinfo -q"
	cmd := exec.Command(ShellToUse, "-c", app)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
		return sensu.CheckStateUnknown, 0, 0, 0, 0, fmt.Errorf("Yum failed: %s", err)
	}

	rCri, _ := regexToNum(stdout.String(), `(?P<num>[0-9]*) Critical Security`)
	rImp, _ := regexToNum(stdout.String(), `(?P<num>[0-9]*) Important Security`)
	rMod, _ := regexToNum(stdout.String(), `(?P<num>[0-9]*) Moderate Security`)
	rBug, _ := regexToNum(stdout.String(), `(?P<num>[0-9]*) Bugfix`)

	//fmt.Printf("stdout\n%s\n", stdout.String())
	//fmt.Printf("stderr\n%s\n", stderr.String())

	numSec := rCri + rImp + rMod
	numPatch := numSec + rBug
	numImp := rImp
	numCrit := rCri

	var retState int = sensu.CheckStateOK
	if numCrit > secCrit && secCrit != -1 {
		retState = sensu.CheckStateCritical
	} else if numSec > secWarn && secWarn != -1 {
		retState = sensu.CheckStateWarning
	}
	return retState, numPatch, numSec, numImp, numCrit, nil
}
