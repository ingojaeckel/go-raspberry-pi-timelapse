// Package files lists, archives, deletes and prunes the photos in the storage
// folder. Every access by name goes through os.Root, so names from HTTP
// requests cannot escape the folder.
package files

import (
	"archive/tar"
	"archive/zip"
	"cmp"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"slices"
	"strings"
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
func ListFiles(dir string) ([]File, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	files := make([]File, 0, len(entries))
	for _, e := range entries {
		if !isPhoto(e) {
			continue
		}
		info, err := e.Info()
		if errors.Is(err, fs.ErrNotExist) {
			continue // deleted since ReadDir
		}
		if err != nil {
			return nil, err
		}
		files = append(files, newFile(info))
	}
	slices.SortFunc(files, oldestFirst)
	return files, nil
}

// isPhoto skips directories and hidden files such as photos still being written.
func isPhoto(e fs.DirEntry) bool {
	return e.Type().IsRegular() && !strings.HasPrefix(e.Name(), ".")
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
	info, err := root.Lstat(name)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("%s: %w", name, ErrNotAFile)
	}
	return root.Remove(name)
}

// WriteTar streams a tar archive of the named files in dir to w.
func WriteTar(w io.Writer, dir string, names []string) error {
	tw := tar.NewWriter(w)
	err := eachFile(dir, names, func(name string, info fs.FileInfo, f io.Reader) error {
		hdr := &tar.Header{Name: name, Size: info.Size(), Mode: int64(info.Mode().Perm()), ModTime: info.ModTime()}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		_, err := io.Copy(tw, f)
		return err
	})
	return errors.Join(err, tw.Close())
}

// WriteZip streams an uncompressed zip archive of the named files in dir to w.
// JPEGs do not compress, so storing them saves CPU on the Pi.
func WriteZip(w io.Writer, dir string, names []string) error {
	zw := zip.NewWriter(w)
	err := eachFile(dir, names, func(name string, info fs.FileInfo, f io.Reader) error {
		entry, err := zw.CreateHeader(&zip.FileHeader{Name: name, Modified: info.ModTime(), Method: zip.Store})
		if err != nil {
			return err
		}
		_, err = io.Copy(entry, f)
		return err
	})
	return errors.Join(err, zw.Close())
}

// eachFile opens each named file inside dir in turn and closes it before
// moving on to the next one.
func eachFile(dir string, names []string, fn func(name string, info fs.FileInfo, f io.Reader) error) error {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	defer root.Close()
	for _, name := range names {
		if err := withFile(root, name, fn); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}
	return nil
}

func withFile(root *os.Root, name string, fn func(string, fs.FileInfo, io.Reader) error) error {
	f, err := root.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	return fn(name, info, f)
}

// Reclaim deletes the oldest photos in dir until at least minFree bytes are
// free on its file system, so the camera keeps the most recent photos like a
// ring buffer. It returns the names of the deleted files, and ErrDiskFull if
// deleting every photo was not enough.
func Reclaim(dir string, minFree uint64) ([]string, error) {
	return reclaim(dir, minFree,
		func() (uint64, error) { return FreeBytes(dir) },
		func(name string) error { return RemoveFile(dir, name) })
}

func reclaim(dir string, minFree uint64, free func() (uint64, error), remove func(string) error) ([]string, error) {
	avail, err := free()
	if err != nil || avail >= minFree {
		return nil, err
	}
	photos, err := ListFiles(dir)
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
