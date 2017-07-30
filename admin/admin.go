package admin

import "os/exec"

func HandleCommand(command string) error {
	if command == "shutdown" {
		return execute([]string{}, "/usr/bin/sudo", "/sbin/shutdown", "-h", "now")
	}
	if command == "restart" {
		return execute([]string{}, "/usr/bin/sudo", "/sbin/shutdown", "-r", "now")
	}
	return nil
}

// RunCommand Execute and return the output to the caller.
func RunCommand(command string, args ...string) (string, error) {
	bytes, err := exec.Command(command, args...).CombinedOutput()
	return string(bytes), err
}
