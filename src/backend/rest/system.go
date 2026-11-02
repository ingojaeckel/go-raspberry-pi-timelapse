package rest

import (
	"context"
	"os/exec"
	"time"
)

// RunCommand runs a system command and returns its combined output. It is
// the production implementation of Server.Run.
func RunCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.WaitDelay = 2 * time.Second
	return cmd.CombinedOutput()
}

// vcgencmdPath finds vcgencmd: /usr/bin on current Raspberry Pi OS,
// /opt/vc/bin on older releases.
func vcgencmdPath() string {
	if p, err := exec.LookPath("vcgencmd"); err == nil {
		return p
	}
	return "/opt/vc/bin/vcgencmd"
}
