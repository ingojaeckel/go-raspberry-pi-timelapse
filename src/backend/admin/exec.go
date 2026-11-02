package admin

import (
	"bufio"
	"log"
	"os/exec"
)

func execute(env []string, command string, args ...string) error {
	log.Printf("execute(env=%v, command=%s, args=%v)\n", env, command, args)

	cmd := exec.Command(command, args...)
	outReader, err := cmd.StdoutPipe()
	errReader, _ := cmd.StderrPipe()
	if err != nil {
		return err
	}

	stdoutScanner := bufio.NewScanner(outReader)
	go func() {
		for stdoutScanner.Scan() {
			log.Printf("[cmd-stdout] %s\n", stdoutScanner.Text())
		}
	}()
	errScanner := bufio.NewScanner(errReader)
	go func() {
		for errScanner.Scan() {
			log.Printf("[cmd-stderr] %s\n", errScanner.Text())
		}
	}()
	return cmd.Run()
}
