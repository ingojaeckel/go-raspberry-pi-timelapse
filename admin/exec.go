package admin

import (
	"bufio"
	"fmt"
	"os/exec"
)

func execute(env []string, command string, args ...string) error {
	fmt.Printf("execute(env=%v, command=%s, args=%v)\n", env, command, args)

	cmd := exec.Command(command, args...)
	outReader, err := cmd.StdoutPipe()
	errReader, _ := cmd.StderrPipe()
	if err != nil {
		return err
	}

	stdoutScanner := bufio.NewScanner(outReader)
	go func() {
		for stdoutScanner.Scan() {
			fmt.Printf("[cmd-stdout] %s\n", stdoutScanner.Text())
		}
	}()
	errScanner := bufio.NewScanner(errReader)
	go func() {
		for errScanner.Scan() {
			fmt.Printf("[cmd-stderr] %s\n", errScanner.Text())
		}
	}()
	return cmd.Run()
}
