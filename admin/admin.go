package admin

import (
	"fmt"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/files"
	"os"
	"os/exec"
)

func HandleCommand(command string) error {
	if command == "shutdown" {
		return execute([]string{}, "/usr/bin/sudo", "/sbin/shutdown", "-h", "now")
	}
	if command == "restart" {
		return execute([]string{}, "/usr/bin/sudo", "/sbin/shutdown", "-r", "now")
	}
	if command == "clear" {
		images, e := files.ListFiles("timelapse-pictures", true)
		if e != nil {
			return e
		}
		for _, f := range images {
			path := "timelapse-pictures/" + f.Name
			if err := os.Remove(path); err != nil {
				return err
			}
			fmt.Println("Removed file " + path)
		}

	}
	return nil
}

// RunCommand Execute and return the output to the caller.
func RunCommand(command string, args ...string) (string, error) {
	bytes, err := exec.Command(command, args...).CombinedOutput()
	return string(bytes), err
}
