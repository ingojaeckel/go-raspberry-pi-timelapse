package admin

import (
	"github.com/facebookgo/ensure"
	"testing"
)

func TestHandleCommand(t *testing.T) {
	ensure.Nil(t, HandleCommand("non-exisinting"))
}

func TestRunCommand(t *testing.T) {
	res, err := RunCommand("ls")
	ensure.Nil(t, err)
	ensure.NotDeepEqual(t, 0, len(res))
}

func TestHandleExecute(t *testing.T) {
	ensure.Nil(t, execute([]string{"KEY=value"}, "/bin/ls", "-la"))
}
