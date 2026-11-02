package files

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"slices"
	"testing"
	"time"
)

// photoDir creates files in a temp dir with increasing modification times.
func photoDir(t *testing.T, names ...string) string {
	t.Helper()
	dir := t.TempDir()
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i, n := range names {
		p := filepath.Join(dir, n)
		if err := os.WriteFile(p, []byte("content of "+n), 0o600); err != nil {
			t.Fatal(err)
		}
		mt := base.Add(time.Duration(i) * time.Minute)
		if err := os.Chtimes(p, mt, mt); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func names(fs []File) []string {
	var out []string
	for _, f := range fs {
		out = append(out, f.Name)
	}
	return out
}

func TestListFilesSortsOldestFirst(t *testing.T) {
	dir := photoDir(t, "b.jpg", "a.jpg", "c.jpg")
	got, err := ListFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"b.jpg", "a.jpg", "c.jpg"}; !slices.Equal(names(got), want) {
		t.Fatalf("got %v, want %v", names(got), want)
	}
	if got[0].Bytes != int64(len("content of b.jpg")) || got[0].ModTimeEpoch == 0 || got[0].ModTime == "" {
		t.Fatalf("missing metadata: %+v", got[0])
	}
}

func TestListFilesSkipsDirectoriesAndHiddenPartialFiles(t *testing.T) {
	dir := photoDir(t, "a.jpg", ".partial-b.jpg")
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o700); err != nil {
		t.Fatal(err)
	}
	got, err := ListFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"a.jpg"}; !slices.Equal(names(got), want) {
		t.Fatalf("got %v, want %v", names(got), want)
	}
}

func TestListFilesMissingDir(t *testing.T) {
	if _, err := ListFiles(filepath.Join(t.TempDir(), "nope")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("err = %v", err)
	}
}

func TestRemoveFile(t *testing.T) {
	dir := photoDir(t, "a.jpg")
	if err := RemoveFile(dir, "a.jpg"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "a.jpg")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("file still exists")
	}
}

func TestRemoveFileRejectsDirectories(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "sub"), 0o700)
	if err := RemoveFile(dir, "sub"); !errors.Is(err, ErrNotAFile) {
		t.Fatalf("err = %v", err)
	}
}

func TestRemoveFileRejectsPathsOutsideTheFolder(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "photos")
	os.Mkdir(dir, 0o700)
	victim := filepath.Join(parent, "victim.txt")
	os.WriteFile(victim, nil, 0o600)
	for _, name := range []string{"../victim.txt", "/etc/passwd", "sub/../../victim.txt"} {
		if err := RemoveFile(dir, name); err == nil {
			t.Errorf("RemoveFile(%q) succeeded", name)
		}
	}
	if _, err := os.Stat(victim); err != nil {
		t.Fatal("file outside the folder was removed")
	}
}

func TestWriteTarUsesBaseNamesAndContent(t *testing.T) {
	dir := photoDir(t, "a.jpg", "b.jpg")
	var buf bytes.Buffer
	if err := WriteTar(&buf, dir, []string{"a.jpg", "b.jpg"}); err != nil {
		t.Fatal(err)
	}
	tr := tar.NewReader(&buf)
	var got []string
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(tr)
		if string(body) != "content of "+hdr.Name {
			t.Errorf("%s has content %q", hdr.Name, body)
		}
		got = append(got, hdr.Name)
	}
	if want := []string{"a.jpg", "b.jpg"}; !slices.Equal(got, want) {
		t.Fatalf("got %v", got)
	}
}

func TestWriteZipRoundTrips(t *testing.T) {
	dir := photoDir(t, "a.jpg", "b.jpg")
	var buf bytes.Buffer
	if err := WriteZip(&buf, dir, []string{"a.jpg", "b.jpg"}); err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	if len(zr.File) != 2 || zr.File[0].Name != "a.jpg" || zr.File[0].Method != zip.Store {
		t.Fatalf("unexpected entries: %+v", zr.File)
	}
	rc, _ := zr.File[1].Open()
	body, _ := io.ReadAll(rc)
	if string(body) != "content of b.jpg" {
		t.Fatalf("content %q", body)
	}
}

func TestWriteArchiveReportsMissingFile(t *testing.T) {
	dir := photoDir(t, "a.jpg")
	for name, write := range map[string]func(io.Writer, string, []string) error{"tar": WriteTar, "zip": WriteZip} {
		if err := write(io.Discard, dir, []string{"a.jpg", "gone.jpg"}); err == nil {
			t.Errorf("%s: expected an error for a missing file", name)
		}
	}
}

// failingWriter simulates a client that disconnects after n bytes.
type failingWriter struct{ n int }

func (w *failingWriter) Write(p []byte) (int, error) {
	if w.n <= 0 {
		return 0, errors.New("client went away")
	}
	w.n -= len(p)
	return len(p), nil
}

func openFDs(t *testing.T) int {
	entries, err := os.ReadDir("/dev/fd")
	if err != nil {
		t.Skip("no /dev/fd on this platform")
	}
	return len(entries)
}

func TestWriteArchiveClosesEveryFile(t *testing.T) {
	var many []string
	for i := range 200 {
		many = append(many, fmt.Sprintf("%03d.jpg", i))
	}
	dir := photoDir(t, many...)
	// Without GC, unclosed files stay open, so a leak shows in the fd count.
	defer debug.SetGCPercent(debug.SetGCPercent(-1))

	for name, write := range map[string]func(io.Writer, string, []string) error{"tar": WriteTar, "zip": WriteZip} {
		before := openFDs(t)
		if err := write(io.Discard, dir, many); err != nil {
			t.Fatal(err)
		}
		if err := write(&failingWriter{n: 2000}, dir, many); err == nil {
			t.Fatalf("%s: expected the write error to be returned", name)
		}
		if after := openFDs(t); after > before {
			t.Errorf("%s leaked %d file descriptors", name, after-before)
		}
	}
}

func TestReclaimDeletesOldestUntilEnoughSpace(t *testing.T) {
	dir := photoDir(t, "1.jpg", "2.jpg", "3.jpg", "4.jpg")
	free := uint64(100)
	freeFn := func() (uint64, error) { return free, nil }
	remove := func(name string) error {
		free += 60 // each photo frees 60 bytes
		return RemoveFile(dir, name)
	}
	removed, err := reclaim(dir, 200, freeFn, remove)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"1.jpg", "2.jpg"}; !slices.Equal(removed, want) {
		t.Fatalf("removed %v, want %v", removed, want)
	}
	left, _ := ListFiles(dir)
	if want := []string{"3.jpg", "4.jpg"}; !slices.Equal(names(left), want) {
		t.Fatalf("left %v", names(left))
	}
}

func TestReclaimDoesNothingWithEnoughSpace(t *testing.T) {
	dir := photoDir(t, "1.jpg")
	removed, err := reclaim(dir, 200, func() (uint64, error) { return 500, nil }, func(string) error {
		t.Fatal("must not delete")
		return nil
	})
	if err != nil || len(removed) != 0 {
		t.Fatalf("removed=%v err=%v", removed, err)
	}
}

func TestReclaimStopsWhenNothingIsLeft(t *testing.T) {
	dir := photoDir(t, "1.jpg")
	removed, err := reclaim(dir, 200, func() (uint64, error) { return 0, nil }, func(n string) error { return RemoveFile(dir, n) })
	if !errors.Is(err, ErrDiskFull) {
		t.Fatalf("err = %v, want ErrDiskFull", err)
	}
	if len(removed) != 1 {
		t.Fatalf("removed %v", removed)
	}
}

func TestFreeBytes(t *testing.T) {
	free, err := FreeBytes(t.TempDir())
	if err != nil {
		t.Skip("free space not supported:", err)
	}
	if free == 0 {
		t.Fatal("expected some free space in the temp dir")
	}
}

func BenchmarkListFiles(b *testing.B) {
	dir := b.TempDir()
	for i := range 1000 {
		os.WriteFile(filepath.Join(dir, fmt.Sprintf("%04d.jpg", i)), nil, 0o600)
	}
	for b.Loop() {
		ListFiles(dir)
	}
}
