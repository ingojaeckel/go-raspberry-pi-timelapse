package admin

func HandleCommand(command string) error {
	if command == "shutdown" {
		return execute([]string{}, "/usr/bin/sudo", "/sbin/shutdown", "-h", "now")
	}
	if command == "restart" {
		return execute([]string{}, "/usr/bin/sudo", "/sbin/shutdown", "-r", "now")
	}
	return nil
}
