package files

import (
	"archive/tar"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/facebookgo/ensure"
)

func TestListFiles(t *testing.T) {
	w, _ := os.Getwd()
	list, e := ListFiles(w, true)
	ensure.Nil(t, e)
	ensure.DeepEqual(t, 2, len(list))
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

func TestCanServeFile(t *testing.T) {
	s1, e1 := CanServeFile("files.go", 10*1024*1024)
	ensure.Nil(t, e1)
	ensure.True(t, s1)

	s2, e2 := CanServeFile("files.go", 1)
	ensure.False(t, s2)
	ensure.Nil(t, e2)

	s3, e3 := CanServeFile("does_not_exist.go", 10*1024*1024)
	ensure.False(t, s3)
	ensure.NotNil(t, e3)
	ensure.True(t, os.IsNotExist(e3))
}

func TestTarTwoFilesWithPipe(t *testing.T) {
	f := []string{"files.go", "files_test.go"}
	pr, pw := io.Pipe()

	go func() {
		err := TarWithPipes(f, pw)
		ensure.Nil(t, err)
		defer pw.Close()
	}()

	// Open the tar archive for reading.
	tr := tar.NewReader(pr)

	count := 0
	// Iterate through the files in the archive.
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // end of tar archive
		}
		ensure.Nil(t, err)
		ensure.DeepEqual(t, f[count], hdr.Name)

		fileContent, _ := ioutil.ReadFile(f[count])
		ensure.DeepEqual(t, int64(len(fileContent)), hdr.Size)

		count++
	}
	ensure.DeepEqual(t, count, 2)
}
