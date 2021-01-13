package ubuntu

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
)

// Severity enum
// List of severities can be found here
// https://askubuntu.com/questions/963803/how-is-the-severity-priority-of-a-vulnerability-in-the-ubuntu-cve-tracker-determ
const (
	sev_neg     int = iota   
	sev_low
	sev_med
	sev_high
	sev_crit
	sev_unknown
)


// Prefered shell for command line execution
const ShellToUse = "bash"

func getSeverity(pkgName string) (int, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	app := "/usr/bin/apt changelog '" + pkgName + "' 2> /dev/null | head -2 | tail -1"
	cmd := exec.Command(ShellToUse, "-c", app)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if (err != nil) {
		log.Printf("error %s: %v\n", pkgName, err)
		return sev_unknown, err
	}
	lineText := stdout.String()
	fmt.Println(lineText)

	// Capture the criticality string from the output
    // ^[a-z]*.*; urgency=([a-z]*)$
	r := regexp.MustCompile(`^.*; urgency=(?P<urgency>(negligible|low|medium|high|critical))`)
	match := r.FindAllStringSubmatch(lineText, -1)
	if len(match) == 0 {
		fmt.Println("stdout: ", lineText)
		return sev_unknown, nil
	}

	names := r.SubexpNames()
	m := map[string]string{}
	for i, n := range match[0] {
		m[names[i]] = n
	}
	urgency := m["urgency"]
	fmt.Println(urgency)

	if urgency == "negligible" {
		return sev_neg, nil
	} else if urgency == "low" {
		return sev_low, nil
	} else if urgency == "medium" {
		return sev_med, nil
	} else if urgency == "high" {
		return sev_high, nil
	} else if urgency == "critical" {
		return sev_crit, nil
	}

	// Severity not found
    err = errors.New(fmt.Sprintf("Severity not found for %s : %s\n", pkgName, stdout.String()))
	return sev_unknown, err
}

// Get the names of the security updates and map them
// ^(?P<name>[^ ]*) (?P<version>[0-9\.]*[^ ]*) (?P<arch>[a-z0-9]*) \[upgradable from: (?P<prevVersion>[0-9\.]*[^ ]*)$
func mapSecUpdates() (map[string]string, error) {
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
		return nil, nil
	}

	r := regexp.MustCompile(`^(?P<name>[^/]*).*? (?P<version>[0-9.]*[^ ]*) (?P<arch>[a-z0-9]*) \[upgradable from: (?P<prevVersion>[0-9.]*[^ ]*)]$`)
	scanner := bufio.NewScanner(bytes.NewReader(stdout.Bytes()))
	for scanner.Scan()  {
		match := r.FindAllStringSubmatch(scanner.Text(), -1)
		if len(match) != 0 {
			patchNames := r.SubexpNames()
			m := make(map[string]string)
			for i, n := range match[0] {
				m[patchNames[i]] = n
			}
			fmt.Println(m)

			sev, _ := getSeverity(m["name"])
			fmt.Println(sev)
		}
	}

	//fmt.Printf("stdout\n%s\n", stdout.String())
	//fmt.Printf("stderr\n%s\n", stderr.String())
	m := map[string]string{}
	return m, nil
}

// checkPatchUbuntu checks if a file exists (...and is not a directory)
// Borrowed from: https://golangcode.com/check-if-a-file-exists/
// bool: Patches applied, int: total updates, int: security updates, int: critical updates
func CheckPatch() (int, int, int, error) {
	// info, err := os.Stat(filename)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// Get the severities
	_, err1 := mapSecUpdates()
	if err1 != nil {
		fmt.Println(fmt.Errorf("Problem with security update info: %s\n", err1))
	}

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

	//fmt.Printf("stdout\n%s\n", stdout.String())
	//fmt.Printf("stderr\n%s\n", stderr.String())

	num_patch := strings.Count(stdout.String(), "upgradable from")
	num_sec := strings.Count(stdout.String(), "-security")

	return num_patch, num_sec, 0, nil
}
