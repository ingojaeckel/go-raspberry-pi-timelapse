package files

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"
	"testing/synctest"
	"time"
)

// settle backdates dir's modification time past racyWindow, as if the last
// change happened a while ago.
func settle(t testing.TB, dir string) {
	t.Helper()
	old := time.Now().Add(-time.Minute)
	if err := os.Chtimes(dir, old, old); err != nil {
		t.Fatal(err)
	}
}

// write creates a photo that was taken age ago.
func write(t testing.TB, dir, name string, age time.Duration) {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(name), 0o600); err != nil {
		t.Fatal(err)
	}
	mt := time.Now().Add(-age)
	os.Chtimes(p, mt, mt)
}

func listNames(t *testing.T, x *Index) []string {
	t.Helper()
	list, err := x.List()
	if err != nil {
		t.Fatal(err)
	}
	return Names(list)
}

func TestIndexListsLikeListFiles(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "b.jpg", 2*time.Hour)
	write(t, dir, "a.jpg", time.Hour)
	write(t, dir, ".partial-c.jpg", 0)
	os.Mkdir(filepath.Join(dir, "sub"), 0o700)
	want, _ := ListFiles(dir)
	got, err := NewIndex(dir).List()
	if err != nil || !slices.Equal(got, want) || len(got) != 2 {
		t.Fatalf("got %v, %v; want %v", got, err, want)
	}
}

// Once a folder time has been seen for longer than racyWindow, an unchanged
// time means an unchanged folder: it is served from memory without reading
// the folder (which is how the test can tell: it hides a change).
func TestIndexTrustsAnUnchangedFolderTimeAfterRacyWindow(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		dir := t.TempDir()
		write(t, dir, "a.jpg", time.Hour)
		info, _ := os.Stat(dir)
		x := NewIndex(dir)
		first, gen1, _ := x.Generation()
		time.Sleep(racyWindow + time.Second)
		x.List() // reads once more, after the window

		write(t, dir, "hidden-change.jpg", 0)
		os.Chtimes(dir, info.ModTime(), info.ModTime())
		second, gen2, _ := x.Generation()
		if &first[0] != &second[0] || gen1 != gen2 || len(second) != 1 {
			t.Fatalf("an unchanged folder must be served from memory, got %v", Names(second))
		}
		if x.Dir() != dir {
			t.Fatalf("Dir() = %q", x.Dir())
		}
	})
}

func TestIndexSeesNewAndDeletedPhotos(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "a.jpg", 3*time.Hour)
	write(t, dir, "b.jpg", 2*time.Hour)
	settle(t, dir)
	x := NewIndex(dir)
	_, gen1, _ := x.Generation()

	write(t, dir, "c.jpg", time.Hour)
	os.Remove(filepath.Join(dir, "a.jpg"))
	settle(t, dir)
	list, gen2, err := x.Generation()
	if err != nil || !slices.Equal(Names(list), []string{"b.jpg", "c.jpg"}) || gen2 == gen1 {
		t.Fatalf("got %v (gen %d -> %d), %v", Names(list), gen1, gen2, err)
	}
}

func TestIndexKeepsTheGenerationWhenOnlyHiddenFilesChange(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "a.jpg", time.Hour)
	settle(t, dir)
	x := NewIndex(dir)
	_, gen1, _ := x.Generation()
	write(t, dir, ".partial-b.jpg", 0) // changes the folder, not the listing
	if _, gen2, _ := x.Generation(); gen2 != gen1 {
		t.Fatalf("generation changed from %d to %d", gen1, gen2)
	}
}

// A change within racyWindow of the previous one may leave the folder's
// modification time unchanged, so a newly seen time is not trusted yet.
func TestIndexRereadsWhileTheFolderTimeIsNew(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "a.jpg", time.Hour)
	recent := time.Now()
	os.Chtimes(dir, recent, recent)
	x := NewIndex(dir)
	x.List()
	write(t, dir, "b.jpg", 0)
	os.Chtimes(dir, recent, recent) // same tick
	if got := listNames(t, x); !slices.Equal(got, []string{"a.jpg", "b.jpg"}) {
		t.Fatalf("got %v", got)
	}
}

func TestIndexOfAMissingFolder(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "photos")
	x := NewIndex(dir)
	if _, err := x.List(); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("err = %v", err)
	}
	os.Mkdir(dir, 0o700)
	write(t, dir, "a.jpg", 0)
	if got := listNames(t, x); !slices.Equal(got, []string{"a.jpg"}) {
		t.Fatalf("got %v", got)
	}
	os.RemoveAll(dir)
	if list, err := x.List(); err == nil || len(list) != 0 {
		t.Fatalf("got %v, %v after the folder was removed", list, err)
	}
}

func TestIndexIsSafeForConcurrentUse(t *testing.T) {
	dir := t.TempDir()
	x := NewIndex(dir)
	var wg sync.WaitGroup
	for i := range 4 {
		wg.Go(func() {
			for j := range 20 {
				write(t, dir, fmt.Sprintf("%d-%02d.jpg", i, j), 0)
				if _, err := x.List(); err != nil {
					t.Error(err)
				}
			}
		})
	}
	wg.Wait()
	if got := listNames(t, x); len(got) != 80 {
		t.Fatalf("listed %d photos, want 80", len(got))
	}
}

// BenchmarkListFiles compares listing a folder from scratch with the Index
// the server uses: unchanged between calls (the gallery polling between two
// captures) and with one new photo per call. 44640 photos are a month at one
// photo per minute.
func BenchmarkListFiles(b *testing.B) {
	for _, n := range []int{1000, 44640} {
		dir := b.TempDir()
		for i := range n {
			os.WriteFile(filepath.Join(dir, fmt.Sprintf("%06d.jpg", i)), nil, 0o600)
		}
		b.Run(fmt.Sprintf("uncached/%d", n), func(b *testing.B) {
			for b.Loop() {
				ListFiles(dir)
			}
		})
		b.Run(fmt.Sprintf("index-unchanged/%d", n), func(b *testing.B) {
			x := NewIndex(dir)
			x.List()
			x.seen = x.seen.Add(-time.Minute) // as if seen a minute ago
			x.List()
			for b.Loop() {
				x.List()
			}
		})
		b.Run(fmt.Sprintf("index-new-photo/%d", n), func(b *testing.B) {
			x := NewIndex(dir)
			x.List()
			i := 0
			for b.Loop() {
				os.WriteFile(filepath.Join(dir, fmt.Sprintf("new-%06d.jpg", i)), nil, 0o600)
				i++
				x.List()
			}
		})
	}
}
