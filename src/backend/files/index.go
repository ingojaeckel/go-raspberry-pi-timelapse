package files

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
)

// racyWindow is how long a folder modification time must have been seen
// before it is trusted: file systems store it with coarse granularity (ext4
// a few milliseconds, FAT two seconds), so a second change within the same
// tick leaves it unchanged. A folder read again more than racyWindow after
// its time was first seen includes every change with that time.
const racyWindow = 3 * time.Second

// Index lists the photos in a folder and remembers them between calls, so
// that the gallery, which polls the listing, does not stat every photo every
// time. A month at one photo per minute is 44640 photos, which takes seconds
// to stat on a Pi Zero.
//
// Photos are written once and never modified (timelapse.Camera renames
// complete files into place and never overwrites one), so a name seen before
// keeps its size and time. List therefore re-reads the folder only when its
// modification time changed (or might hide a change, see racyWindow), and
// then stats only the names it has not seen. Elapsed time is measured with
// the monotonic clock, so wall clock jumps (NTP, a Pi without RTC) do not
// matter. It is safe for concurrent use.
type Index struct {
	dir string

	mu       sync.Mutex
	byName   []File    // as last read, sorted by name like os.ReadDir
	sorted   []File    // byName oldest first; handed out, never modified
	dirMod   time.Time // modification time of dir when byName was read
	seen     time.Time // when dirMod was first seen (monotonic)
	loaded   bool      // byName and sorted hold a listing
	verified bool      // read again after racyWindow: unchanged dirMod means unchanged
	gen      uint64    // incremented whenever the listing changes
}

// NewIndex returns an index of the photos in dir.
func NewIndex(dir string) *Index { return &Index{dir: dir} }

// Dir is the folder the index lists.
func (x *Index) Dir() string { return x.dir }

// List returns the photos oldest first, like ListFiles. The slice is shared
// between callers and must not be modified.
func (x *Index) List() ([]File, error) {
	list, _, err := x.Generation()
	return list, err
}

// Generation is List plus a number that changes whenever the listing
// changes, so that callers can cache what they derive from it.
func (x *Index) Generation() ([]File, uint64, error) {
	x.mu.Lock()
	defer x.mu.Unlock()
	info, err := os.Stat(x.dir)
	if err != nil {
		x.byName, x.sorted, x.loaded, x.verified = nil, nil, false, false
		x.gen++
		return nil, x.gen, err
	}
	if x.verified && info.ModTime().Equal(x.dirMod) {
		return x.sorted, x.gen, nil
	}
	return x.refresh(info.ModTime())
}

// refresh re-reads the folder, reusing what it knows about each name.
func (x *Index) refresh(dirMod time.Time) ([]File, uint64, error) {
	started := time.Now()
	names, err := readNames(x.dir)
	if err != nil {
		return nil, x.gen, err
	}
	byName, changed, err := merge(x.dir, x.byName, names)
	if err != nil {
		return nil, x.gen, err
	}
	if !x.loaded || !dirMod.Equal(x.dirMod) {
		x.dirMod, x.seen, x.verified = dirMod, started, false
	} else if started.Sub(x.seen) > racyWindow {
		x.verified = true
	}
	if changed || !x.loaded {
		x.loaded = true
		x.byName = byName
		x.sorted = slices.Clone(byName)
		// Names are capture times, so this is nearly sorted already, which
		// pdqsort handles in close to linear time.
		slices.SortFunc(x.sorted, oldestFirst)
		x.gen++
	}
	return x.sorted, x.gen, nil
}

// readNames returns the names in dir, sorted, without hidden ones. It is
// cheaper than os.ReadDir, which allocates an entry per name.
func readNames(dir string) ([]string, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	names, err := f.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	names = slices.DeleteFunc(names, func(n string) bool { return strings.HasPrefix(n, ".") })
	slices.Sort(names)
	return names, nil
}

// merge lists the photos among names (sorted), taking each known photo from
// old (also sorted by name) and stat'ing only the others: hidden files such
// as photos still being written are already gone, directories and other
// non-regular files are skipped. changed reports whether the result differs
// from old.
func merge(dir string, old []File, names []string) (_ []File, changed bool, _ error) {
	if old != nil && sameNames(old, names) {
		return old, false, nil
	}
	files := make([]File, 0, len(names))
	i := 0
	for _, name := range names {
		for i < len(old) && old[i].Name < name {
			i++ // deleted
			changed = true
		}
		if i < len(old) && old[i].Name == name {
			files = append(files, old[i])
			i++
			continue
		}
		info, err := os.Lstat(filepath.Join(dir, name))
		if errors.Is(err, fs.ErrNotExist) || err == nil && !info.Mode().IsRegular() {
			continue // deleted since reading the names, or not a photo
		}
		if err != nil {
			return nil, false, err
		}
		files = append(files, newFile(info))
		changed = true
	}
	return files, changed || i < len(old), nil
}

func sameNames(files []File, names []string) bool {
	if len(files) != len(names) {
		return false
	}
	for i, f := range files {
		if f.Name != names[i] {
			return false
		}
	}
	return true
}
