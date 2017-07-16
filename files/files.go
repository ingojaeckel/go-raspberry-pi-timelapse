package files

import (
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

func CanServeFile(path string, maxFileSizeBytes int64) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return !fileInfo.IsDir() && fileInfo.Size() <= maxFileSizeBytes, nil
}
