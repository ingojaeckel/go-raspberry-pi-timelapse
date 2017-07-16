package files

import "io/ioutil"

type File struct {
	Name string
}

func ListFiles(dirname string) ([]File, error) {
	fileInfo, e := ioutil.ReadDir(dirname)
	if e != nil {
		return []File{}, e
	}
	files := make([]File, len(fileInfo))
	for i, f := range fileInfo {
		files[i] = File{f.Name()}
	}
	return files, nil
}

func GetFile(path string) ([]byte, error) {
	// TODO add prefix to path if necessary
	return ioutil.ReadFile(path)
}
