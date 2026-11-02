// Package logfile provides a size-capped log file for unattended operation:
// when the file reaches its limit it is renamed to <path>.1 (replacing the
// previous backup) and a new file is started, so logs never use more than
// twice the limit on the SD card.
package logfile

import (
	"bytes"
	"io"
	"os"
	"sync"
)

// Writer is an io.Writer that rotates the file at path. It is safe for
// concurrent use.
type Writer struct {
	mu      sync.Mutex
	path    string
	maxSize int64
	f       *os.File
	size    int64
}

// Open opens path for appending, creating it if needed.
func Open(path string, maxSize int64) (*Writer, error) {
	w := &Writer{path: path, maxSize: maxSize}
	if err := w.open(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *Writer) open() error {
	f, err := os.OpenFile(w.path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return err
	}
	w.f, w.size = f, info.Size()
	return nil
}

func (w *Writer) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.size > 0 && w.size+int64(len(p)) > w.maxSize {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}
	n, err := w.f.Write(p)
	w.size += int64(n)
	return n, err
}

func (w *Writer) rotate() error {
	if err := w.f.Close(); err != nil {
		return err
	}
	if err := os.Rename(w.path, w.path+".1"); err != nil {
		return err
	}
	return w.open()
}

// Close closes the current file.
func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.f.Close()
}

// Tail returns at most the last n bytes of the file at path, starting at the
// beginning of a line.
func Tail(path string, n int64) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return "", err
	}
	start := max(info.Size()-n, 0)
	b, err := io.ReadAll(io.NewSectionReader(f, start, info.Size()-start))
	if err != nil {
		return "", err
	}
	if start > 0 {
		if i := bytes.IndexByte(b, '\n'); i >= 0 {
			b = b[i+1:]
		}
	}
	return string(b), nil
}
