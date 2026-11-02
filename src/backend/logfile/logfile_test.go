package logfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriterAppendsToExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	os.WriteFile(path, []byte("old\n"), 0o644)
	w, err := Open(path, 1024)
	if err != nil {
		t.Fatal(err)
	}
	w.Write([]byte("new\n"))
	w.Close()
	if got, _ := os.ReadFile(path); string(got) != "old\nnew\n" {
		t.Fatalf("got %q", got)
	}
}

func TestWriterRotatesWhenFull(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	w, err := Open(path, 10)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	w.Write([]byte("12345678\n")) // 9 bytes, fits
	w.Write([]byte("abc\n"))      // would exceed 10: rotate first
	if got, _ := os.ReadFile(path + ".1"); string(got) != "12345678\n" {
		t.Fatalf("rotated file = %q", got)
	}
	if got, _ := os.ReadFile(path); string(got) != "abc\n" {
		t.Fatalf("current file = %q", got)
	}
	w.Write([]byte("defghijk\n")) // rotates again, replacing the old backup
	if got, _ := os.ReadFile(path + ".1"); string(got) != "abc\n" {
		t.Fatalf("second rotation = %q", got)
	}
}

func TestWriterNeverExceedsTwiceTheLimitOnDisk(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	w, _ := Open(path, 100)
	defer w.Close()
	for range 1000 {
		w.Write([]byte("0123456789\n"))
	}
	for _, p := range []string{path, path + ".1"} {
		if info, err := os.Stat(p); err != nil || info.Size() > 100 {
			t.Fatalf("%s: size %v err %v", p, info.Size(), err)
		}
	}
}

func TestTailReturnsWholeSmallFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	os.WriteFile(path, []byte("a\nb\n"), 0o644)
	got, err := Tail(path, 1024)
	if err != nil || got != "a\nb\n" {
		t.Fatalf("got %q err %v", got, err)
	}
}

func TestTailStartsAtALineBoundary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	os.WriteFile(path, []byte(strings.Repeat("x", 100)+"\nlast line\n"), 0o644)
	got, err := Tail(path, 15)
	if err != nil || got != "last line\n" {
		t.Fatalf("got %q err %v", got, err)
	}
}

func TestTailMissingFile(t *testing.T) {
	if _, err := Tail(filepath.Join(t.TempDir(), "nope"), 10); err == nil {
		t.Fatal("expected an error")
	}
}

func TestOpenFailsInMissingDirectory(t *testing.T) {
	if _, err := Open(filepath.Join(t.TempDir(), "nope", "app.log"), 10); err == nil {
		t.Fatal("expected an error")
	}
}

func TestWriteFailsWhenRotationFails(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	w, err := Open(path, 5)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	w.Write([]byte("1234\n"))
	os.Mkdir(path+".1", 0o700)                              // the backup name is taken by a directory
	os.WriteFile(filepath.Join(path+".1", "x"), nil, 0o600) // and it is not empty, so rename fails
	if _, err := w.Write([]byte("more\n")); err == nil {
		t.Fatal("expected the rotation error")
	}
}

func TestTailWithoutNewlineInTheWindow(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	os.WriteFile(path, []byte(strings.Repeat("x", 100)), 0o644)
	got, err := Tail(path, 10)
	if err != nil || got != strings.Repeat("x", 10) {
		t.Fatalf("got %q err %v", got, err)
	}
}

func TestWriteAfterCloseFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	w, _ := Open(path, 5)
	w.Write([]byte("1234\n"))
	w.Close()
	if _, err := w.Write([]byte("rotate now\n")); err == nil {
		t.Fatal("expected an error writing to a closed log")
	}
}
