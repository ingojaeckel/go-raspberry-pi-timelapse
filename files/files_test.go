package files

import (
	"archive/tar"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
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

// Benchmarks for performance-sensitive operations

func BenchmarkListFiles(b *testing.B) {
	w, _ := os.Getwd()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ListFiles(w, true)
	}
}

func BenchmarkByAgeSort(b *testing.B) {
	// Create sample file list with 100 files
	files := make([]File, 100)
	for i := 0; i < 100; i++ {
		files[i] = File{
			Name:         "file" + strconv.Itoa(i) + ".jpg",
			ModTimeEpoch: int64(100 - i), // Reverse order
			Bytes:        1024 * int64(i),
		}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Copy to avoid sorting already sorted data
		testFiles := make([]File, len(files))
		copy(testFiles, files)
		sort.Sort(ByAge(testFiles))
	}
}

func BenchmarkByAgeSortFull(b *testing.B) {
	// Create sample file list with 1000 files to test scalability
	files := make([]File, 1000)
	for i := 0; i < 1000; i++ {
		files[i] = File{
			Name:         "file" + strconv.Itoa(i) + ".jpg",
			ModTimeEpoch: int64(1000 - i), // Reverse order
			Bytes:        1024 * int64(i),
		}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Copy to avoid sorting already sorted data
		testFiles := make([]File, len(files))
		copy(testFiles, files)
		sort.Sort(ByAge(testFiles))
	}
}
