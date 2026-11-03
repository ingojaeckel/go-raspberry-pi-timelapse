package files

import (
	"errors"
	"os"
	"path/filepath"
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

func TestReclaimDeletesOldestUntilEnoughSpace(t *testing.T) {
	dir := photoDir(t, "1.jpg", "2.jpg", "3.jpg", "4.jpg")
	free := uint64(100)
	freeFn := func() (uint64, error) { return free, nil }
	remove := func(name string) error {
		free += 60 // each photo frees 60 bytes
		return RemoveFile(dir, name)
	}
	removed, err := reclaim(NewIndex(dir).List, 200, freeFn, remove)
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
	removed, err := reclaim(NewIndex(dir).List, 200, func() (uint64, error) { return 500, nil }, func(string) error {
		t.Fatal("must not delete")
		return nil
	})
	if err != nil || len(removed) != 0 {
		t.Fatalf("removed=%v err=%v", removed, err)
	}
}

func TestReclaimStopsWhenNothingIsLeft(t *testing.T) {
	dir := photoDir(t, "1.jpg")
	removed, err := reclaim(NewIndex(dir).List, 200, func() (uint64, error) { return 0, nil }, func(n string) error { return RemoveFile(dir, n) })
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

func TestNames(t *testing.T) {
	got := Names([]File{{Name: "a.jpg"}, {Name: "b.jpg"}})
	if !slices.Equal(got, []string{"a.jpg", "b.jpg"}) {
		t.Fatalf("got %v", got)
	}
	if got := Names(nil); len(got) != 0 {
		t.Fatalf("got %v", got)
	}
}

func TestRemoveFileMissing(t *testing.T) {
	if err := RemoveFile(t.TempDir(), "nope.jpg"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("err = %v", err)
	}
	if err := RemoveFile(filepath.Join(t.TempDir(), "no-dir"), "a.jpg"); err == nil {
		t.Fatal("expected an error for a missing folder")
	}
}

func TestReclaimUsesRealFreeSpace(t *testing.T) {
	dir := photoDir(t, "1.jpg", "2.jpg")
	if removed, err := NewIndex(dir).Reclaim(0, func() (uint64, error) { return FreeBytes(dir) }); err != nil || len(removed) != 0 {
		t.Fatalf("nothing to do: removed %v err %v", removed, err)
	}
	removed, err := NewIndex(dir).Reclaim(1<<62, func() (uint64, error) { return FreeBytes(dir) }) // more than any disk
	if !errors.Is(err, ErrDiskFull) || len(removed) != 2 {
		t.Fatalf("removed %v err %v", removed, err)
	}
}

func TestReclaimReportsErrors(t *testing.T) {
	dir := photoDir(t, "1.jpg")
	boom := errors.New("boom")
	full := func() (uint64, error) { return 0, nil }

	if _, err := reclaim(NewIndex(dir).List, 1, func() (uint64, error) { return 0, boom }, nil); err != boom {
		t.Errorf("free space error: %v", err)
	}
	if _, err := reclaim(NewIndex(filepath.Join(dir, "nope")).List, 1, full, nil); err == nil {
		t.Error("listing error expected")
	}
	if _, err := reclaim(NewIndex(dir).List, 1, full, func(string) error { return boom }); err != boom {
		t.Errorf("remove error: %v", err)
	}
}

func TestFreeBytesMissingDir(t *testing.T) {
	if _, err := FreeBytes(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Fatal("expected an error")
	}
}
