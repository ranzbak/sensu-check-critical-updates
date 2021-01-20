package pendingtime

import (
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/sensu-community/sensu-plugin-sdk/sensu"
)

// GetLastAccess Given a path, the time in seconds since the file was last updated is returned
func GetLastAccess(path string) (int64, error){
        fileinfo, err := os.Stat(path)
        if err != nil {
            log.Fatal(err)
            return 0, fmt.Errorf("File %s not found", path) 
        }
        atime := time.Now().Unix() - fileinfo.ModTime().Unix()
        return atime, nil
}

// TouchLastAccess touch the witness file to track the time
func TouchLastAccess(path string) (error){
        af, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
        if err != nil {
                return fmt.Errorf("can not open file for writing %s: %s", path, err)
        }
        defer af.Close()

        // Truncate the file to write the new date
        err = af.Truncate(0)
        if err != nil {
                return fmt.Errorf("truncate failed on %s: %s", path, err)
        }

        _, err = fmt.Fprintf(af, "%s", time.Now().UTC().Format(time.UnixDate))
        if err != nil {
                return fmt.Errorf("writing time failed: %s", err)
        }

        return nil
}

// CreateIfNotExist if the file is not present create it
func CreateIfNotExist(path string) (error) {
        _, err := os.Stat(path)
        if os.IsNotExist(err) {
                err2 := TouchLastAccess(path)
                if err2 != nil {
                        return err2
                }
        }
        return nil
}

// PendingTime This function returns the status, and the days since the patch counter was last 0
// status, days since no patches, error
func PendingTime(path string, pending int, daysWarn int, daysCrit int) (int, int, error){
        // Create the empty file if it doesn't exist yet
        err := CreateIfNotExist(path)
        if err != nil {
                return 0, 0, err
        }

        // Only when no patches are outstanding touch the file
        if pending == 0 {
                err := TouchLastAccess(path)
                if err != nil {
                        return 0, 0, err
                }
        }
        // Check the file last mod time to check last patch time
        atime, err := GetLastAccess(path)    
        if err != nil {
                return 0, 0, err
        }

        // Number of days since patch
        daysSincePatch := int(math.Floor(float64(atime)/(24*60*60)))

        // Check if the last time the amount of patches is too long ago
        var retState int = sensu.CheckStateOK
        if daysSincePatch > daysCrit && daysCrit != -1 {
                retState = sensu.CheckStateCritical
        } else if daysSincePatch > daysWarn && daysWarn != -1 {
                retState = sensu.CheckStateWarning
        }

        return retState, daysSincePatch, nil
}
