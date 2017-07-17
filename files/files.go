package files

import (
	"archive/tar"
	"bytes"
	"io/ioutil"
	"os"
)

type File struct {
	Name    string `json:"name"`
	ModTime string `json:"mod_time"`
	IsDir   bool   `json:"is_dir"`
}

func ListFiles(dirname string) ([]File, error) {
	fileInfo, e := ioutil.ReadDir(dirname)
	if e != nil {
		return []File{}, e
	}
	files := make([]File, len(fileInfo))
	for i, f := range fileInfo {
		files[i] = File{
			Name:    f.Name(),
			ModTime: f.ModTime().String(),
			IsDir:   f.IsDir(),
		}
	}
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
