package files

import (
	"archive/tar"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"sort"
)

type File struct {
	Name         string `json:"name"`
	ModTime      string `json:"mod_time"`
	ModTimeEpoch int64  `json:"mod_time_epoch"`
	IsDir        bool   `json:"is_dir"`
	Bytes        int64  `json:"bytes"`
}

func ListFiles(dirname string, skipDirectories bool) ([]File, error) {
	fileInfo, e := ioutil.ReadDir(dirname)
	if e != nil {
		return []File{}, e
	}
	numberOfFiles := 0
	files := make([]File, len(fileInfo))
	for _, f := range fileInfo {
		if skipDirectories && f.IsDir() {
			continue
		}
		files[numberOfFiles] = File{
			Name:         f.Name(),
			ModTime:      f.ModTime().String(),
			ModTimeEpoch: f.ModTime().Unix(),
			IsDir:        f.IsDir(),
			Bytes:        f.Size(),
		}
		numberOfFiles = numberOfFiles + 1
	}

	files = files[:numberOfFiles]
	sort.Sort(ByAge(files))

	return files, nil
}

func GetFile(path string) ([]byte, error) {
	// TODO add prefix to path if necessary
	return ioutil.ReadFile(path)
}

// CanServeFile True if the given file can be served via HTTP. False otherwise because it does not exist, is a directory, or because it is too large.
func CanServeFile(path string, maxFileSizeBytes int64) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return !fileInfo.IsDir() && fileInfo.Size() <= maxFileSizeBytes, nil
}

// Tar creates a .tar archive containing all files within the filePaths array.
func Tar(filePaths []string) ([]byte, error) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	for _, f := range filePaths {
		info, err := os.Stat(f)
		if err != nil {
			return []byte{}, err
		}
		// TODO consider using tar.FileInfoHeader(info, link)
		if err := tw.WriteHeader(&tar.Header{Name: info.Name(), Mode: 0400, Size: info.Size()}); err != nil {
			return []byte{}, err
		}
		content, err := ioutil.ReadFile(f)
		if err != nil {
			return []byte{}, err
		}
		if _, err := tw.Write(content); err != nil {
			return []byte{}, err
		}
	}
	if err := tw.Close(); err != nil {
		return []byte{}, err
	}
	// Return in-memory representation of .tar archive containing all files.
	return buf.Bytes(), nil
}

func TarWithPipes(filePaths []string, pw *io.PipeWriter) error {
	tw := tar.NewWriter(pw)

	for _, f := range filePaths {
		info, err := os.Stat(f)
		if err != nil {
			return err
		}
		// TODO consider using tar.FileInfoHeader(info, link)
		if err := tw.WriteHeader(&tar.Header{Name: info.Name(), Mode: 0400, Size: info.Size()}); err != nil {
			return err
		}
		content, err := ioutil.ReadFile(f)
		if err != nil {
			return err
		}
		if _, err := tw.Write(content); err != nil {
			return err
		}
	}
	if err := tw.Close(); err != nil {
		return err
	}
	return nil
}

// BySize Implements the sort.Interface to allow sorting files by their age.
type ByAge []File

func (b ByAge) Len() int {
	return len(b)
}
func (a ByAge) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a ByAge) Less(i, j int) bool {
	return a[i].ModTimeEpoch < a[j].ModTimeEpoch
}
