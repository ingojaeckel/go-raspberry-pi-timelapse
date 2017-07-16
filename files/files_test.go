package files

import (
	"github.com/facebookgo/ensure"
	"os"
	"strings"
	"testing"
)

func TestListFiles(t *testing.T) {
	w, _ := os.Getwd()
	list, e := ListFiles(w)
	ensure.Nil(t, e)
	ensure.DeepEqual(t, 2, len(list))
	ensure.DeepEqual(t, []File{{"files.go"}, {"files_test.go"}}, list)
}

func TestGetFile(t *testing.T) {
	content, e := GetFile("files.go")
	ensure.Nil(t, e)
	ensure.True(t, strings.HasPrefix(string(content), "package files"))
}

func TestGetFile2(t *testing.T) {
	_, e := GetFile("files_does_not_exist.go")
	ensure.NotNil(t, e)
}
