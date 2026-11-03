// Package files lists, archives, deletes and prunes the photos in the storage
// folder. Every access by name goes through os.Root, so names from HTTP
// requests cannot escape the folder.
package files

import (
	"cmp"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

var (
	ErrNotAFile = errors.New("not a regular file")
	ErrDiskFull = errors.New("not enough free disk space, even after deleting every photo")
)

type File struct {
	Name         string `json:"name"`
	ModTime      string `json:"mod_time"`
	ModTimeEpoch int64  `json:"mod_time_epoch"`
	IsDir        bool   `json:"is_dir"`
	Bytes        int64  `json:"bytes"`
}

// ListFiles returns the regular files in dir, oldest first. Hidden files
// (such as photos still being written, see timelapse.Camera) are skipped.
// Use an Index to list the same folder repeatedly.
func ListFiles(dir string) ([]File, error) {
	return NewIndex(dir).List()
}

func newFile(info fs.FileInfo) File {
	return File{
		Name:         info.Name(),
		ModTime:      info.ModTime().String(),
		ModTimeEpoch: info.ModTime().Unix(),
		Bytes:        info.Size(),
	}
}

// oldestFirst orders by modification time, then name.
func oldestFirst(a, b File) int {
	return cmp.Or(cmp.Compare(a.ModTimeEpoch, b.ModTimeEpoch), cmp.Compare(a.Name, b.Name))
}

// Names returns the names of files in order.
func Names(files []File) []string {
	names := make([]string, len(files))
	for i, f := range files {
		names[i] = f.Name
	}
	return names
}

// RemoveFile deletes the regular file name inside dir.
func RemoveFile(dir, name string) error {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	defer root.Close()
	return removeIn(root, name)
}

// RemoveFiles deletes the named regular files inside dir and returns how
// many it deleted. failed is called for each name that could not be deleted.
func RemoveFiles(dir string, names []string, failed func(name string, err error)) (int, error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return 0, err
	}
	defer root.Close()
	removed := 0
	for _, name := range names {
		if err := removeIn(root, name); err != nil {
			failed(name, err)
			continue
		}
		removed++
	}
	return removed, nil
}

func removeIn(root *os.Root, name string) error {
	info, err := root.Lstat(name)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("%s: %w", name, ErrNotAFile)
	}
	return root.Remove(name)
}

// Reclaim deletes the oldest photos in the folder until at least minFree
// bytes are free on its file system, so the camera keeps the most recent
// photos like a ring buffer. It returns the names of the deleted files, and
// ErrDiskFull if deleting every photo was not enough.
// free reports the free bytes, normally FreeBytes of the folder.
func (x *Index) Reclaim(minFree uint64, free func() (uint64, error)) ([]string, error) {
	return reclaim(x.List, minFree, free, func(name string) error { return RemoveFile(x.dir, name) })
}

func reclaim(list func() ([]File, error), minFree uint64, free func() (uint64, error), remove func(string) error) ([]string, error) {
	avail, err := free()
	if err != nil || avail >= minFree {
		return nil, err
	}
	photos, err := list()
	if err != nil {
		return nil, err
	}
	var removed []string
	for _, p := range photos {
		if err := remove(p.Name); err != nil {
			return removed, err
		}
		removed = append(removed, p.Name)
		if avail, err = free(); err != nil || avail >= minFree {
			return removed, err
		}
	}
	return removed, ErrDiskFull
}
