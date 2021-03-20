package files

import (
	"archive/tar"
	"archive/zip"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"sort"
)

var errCannotRemoveDirectory = errors.New("cannot remove a directory")

type File struct {
	Name         string `json:"name"`
	ModTime      string `json:"mod_time"`
	ModTimeEpoch int64  `json:"mod_time_epoch"`
	IsDir        bool   `json:"is_dir"`
	Bytes        int64  `json:"bytes"`
}

func RemoveFile(path string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}
	if fileInfo.IsDir() {
		return errCannotRemoveDirectory
	}
	return os.Remove(path)
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

// TarWithPipes combines all files specified by filePaths.
// Tries to minimize memory usage by using pipes.
// As a result this can only write as quickly as the content is being read.
func TarWithPipes(filePaths []string, pw *io.PipeWriter) error {
	tw := tar.NewWriter(pw)

	for _, f := range filePaths {
		info, err := os.Stat(f)
		if err != nil {
			return err
		}
		hdr, _ := tar.FileInfoHeader(info, info.Name())
		if err = tw.WriteHeader(hdr); err != nil {
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

// ZipWithPipes combines all files specified by filePaths.
// Tries to minimize memory usage by using pipes.
// As a result this can only write as quickly as the content is being read.
func ZipWithPipes(filePaths []string, pw *io.PipeWriter) error {
	w := zip.NewWriter(pw)

	for _, f := range filePaths {
		info, err := os.Stat(f)
		if err != nil {
			return err
		}
		zipFile, err := w.CreateHeader(&zip.FileHeader{
			Name:   info.Name(),
			Method: zip.Store,
		})
		if err != nil {
			return err
		}
		content, err := ioutil.ReadFile(f)
		if err != nil {
			return err
		}
		_, err = zipFile.Write(content)
		if err != nil {
			return err
		}
	}
	if err := w.Close(); err != nil {
		return err
	}
	return nil
}

// ByAge Implements the sort.Interface to allow sorting files by their age.
type ByAge []File

func (b ByAge) Len() int {
	return len(b)
}
func (b ByAge) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}
func (b ByAge) Less(i, j int) bool {
	return b[i].ModTimeEpoch < b[j].ModTimeEpoch
}
