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
	"strings"
	"testing"
	"time"
)

// archiveDir writes photos of different sizes (to exercise tar padding),
// including a name too long for a plain tar header.
func archiveDir(t testing.TB) (string, []File) {
	t.Helper()
	dir := t.TempDir()
	base := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	names := []string{"20260301-120000.jpg", "20260301-120100.jpg", strings.Repeat("long", 30) + ".jpg", "empty.jpg"}
	for i, n := range names {
		content := bytes.Repeat([]byte{byte('a' + i)}, (i*700)%1500)
		p := filepath.Join(dir, n)
		if err := os.WriteFile(p, content, 0o600); err != nil {
			t.Fatal(err)
		}
		mt := base.Add(time.Duration(i) * time.Minute)
		os.Chtimes(p, mt, mt)
	}
	list, err := ListFiles(dir)
	if err != nil || len(list) != len(names) {
		t.Fatalf("listed %v, %v", list, err)
	}
	return dir, list
}

// stdlibTar is what archive/tar writes for the photos.
func stdlibTar(t *testing.T, dir string, photos []File) []byte {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for _, p := range photos {
		content, _ := os.ReadFile(filepath.Join(dir, p.Name))
		tw.WriteHeader(&tar.Header{Typeflag: tar.TypeReg, Name: p.Name, Size: p.Bytes, Mode: 0o600, ModTime: time.Unix(p.ModTimeEpoch, 0)})
		tw.Write(content)
	}
	tw.Close()
	return buf.Bytes()
}

// stdlibZip is what zip.Writer.CreateHeader writes for stored photos.
func stdlibZip(t *testing.T, dir string, photos []File) []byte {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, p := range photos {
		content, _ := os.ReadFile(filepath.Join(dir, p.Name))
		w, err := zw.CreateHeader(&zip.FileHeader{Name: p.Name, Method: zip.Store, Modified: time.Unix(p.ModTimeEpoch, 0)})
		if err != nil {
			t.Fatal(err)
		}
		w.Write(content)
	}
	zw.Close()
	return buf.Bytes()
}

func TestWriteTarMatchesArchiveTarAndTarSize(t *testing.T) {
	dir, photos := archiveDir(t)
	var buf bytes.Buffer
	if err := WriteTar(&buf, dir, photos); err != nil {
		t.Fatal(err)
	}
	if want := stdlibTar(t, dir, photos); !bytes.Equal(buf.Bytes(), want) {
		t.Fatalf("WriteTar wrote %d bytes that differ from archive/tar's %d", buf.Len(), len(want))
	}
	if size, err := TarSize(photos); err != nil || size != int64(buf.Len()) {
		t.Fatalf("TarSize = %d, %v; wrote %d", size, err, buf.Len())
	}
}

func TestWriteZipMatchesZipWriterAndZipSize(t *testing.T) {
	dir, photos := archiveDir(t)
	var buf bytes.Buffer
	if err := WriteZip(&buf, dir, photos); err != nil {
		t.Fatal(err)
	}
	if want := stdlibZip(t, dir, photos); !bytes.Equal(buf.Bytes(), want) {
		t.Fatalf("WriteZip wrote %d bytes that differ from zip.Writer's %d", buf.Len(), len(want))
	}
	if size, err := ZipSize(photos); err != nil || size != int64(buf.Len()) {
		t.Fatalf("ZipSize = %d, %v; wrote %d", size, err, buf.Len())
	}
	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	for i, f := range zr.File {
		rc, _ := f.Open()
		body, err := io.ReadAll(rc) // checks the CRC
		if err != nil || int64(len(body)) != photos[i].Bytes || f.Modified.Unix() != photos[i].ModTimeEpoch {
			t.Fatalf("%s: %d bytes, modified %v, err %v", f.Name, len(body), f.Modified, err)
		}
	}
}

func TestZipFlagsNonASCIINamesAsUTF8(t *testing.T) {
	if fh := zipHeader(File{Name: "café.jpg"}); fh.Flags&0x800 == 0 {
		t.Fatal("expected the UTF-8 flag")
	}
	if fh := zipHeader(File{Name: "cafe.jpg"}); fh.Flags&0x800 != 0 {
		t.Fatal("unexpected UTF-8 flag")
	}
	if d, _ := msDosTime(time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)); d != 1<<5|1 {
		t.Fatalf("dates before 1980 must be clamped, got %#x", d)
	}
}

// More than 65535 entries need zip64 records, which two months at one photo
// per minute reach.
func TestZip64ForManyPhotos(t *testing.T) {
	if testing.Short() {
		t.Skip("writes 70000 files")
	}
	dir := t.TempDir()
	photos := make([]File, 70000)
	for i := range photos {
		photos[i] = File{Name: fmt.Sprintf("%06d.jpg", i), ModTimeEpoch: 1.7e9}
		os.WriteFile(filepath.Join(dir, photos[i].Name), nil, 0o600)
	}
	var buf bytes.Buffer
	if err := WriteZip(&buf, dir, photos); err != nil {
		t.Fatal(err)
	}
	if size, _ := ZipSize(photos); size != int64(buf.Len()) {
		t.Fatalf("ZipSize = %d; wrote %d", size, buf.Len())
	}
	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil || len(zr.File) != len(photos) {
		t.Fatalf("read back %v entries, %v", len(zr.File), err)
	}
}

// fileReceiver records whether it was handed the photo itself, which is what
// net/http needs to use sendfile(2).
type fileReceiver struct {
	bytes.Buffer
	files int
}

func (r *fileReceiver) ReadFrom(src io.Reader) (int64, error) {
	if lr, ok := src.(*io.LimitedReader); ok {
		if _, ok := lr.R.(*os.File); ok {
			r.files++
		}
	}
	return r.Buffer.ReadFrom(src)
}

func TestWriteTarHandsPhotosToReaderFrom(t *testing.T) {
	dir, photos := archiveDir(t)
	var r fileReceiver
	if err := WriteTar(&r, dir, photos); err != nil {
		t.Fatal(err)
	}
	if r.files != len(photos) {
		t.Fatalf("%d of %d photos were handed over as files", r.files, len(photos))
	}
}

type archiver func(io.Writer, string, []File) error

var archivers = map[string]archiver{"tar": WriteTar, "zip": WriteZip}

func TestWriteArchiveFailsWhenAPhotoChanged(t *testing.T) {
	dir, photos := archiveDir(t)
	for name, write := range archivers {
		grown := append([]File(nil), photos...)
		grown[1].Bytes++
		if err := write(io.Discard, dir, grown); !errors.Is(err, ErrChanged) {
			t.Errorf("%s: grown photo: err = %v", name, err)
		}
		gone := append(append([]File(nil), photos...), File{Name: "gone.jpg"})
		if err := write(io.Discard, dir, gone); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("%s: missing photo: err = %v", name, err)
		}
		if err := write(io.Discard, filepath.Join(dir, "nope"), photos); err == nil {
			t.Errorf("%s: expected an error for a missing folder", name)
		}
	}
}

func TestWriteArchiveRejectsDirectoriesAsEntries(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "sub"), 0o700)
	for name, write := range archivers {
		if err := write(io.Discard, dir, []File{{Name: "sub"}}); !errors.Is(err, ErrNotAFile) {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}

func TestArchiveSizeRejectsInvalidNames(t *testing.T) {
	if _, err := TarSize([]File{{Name: "a\x00.jpg"}}); err == nil {
		t.Error("tar: expected an error")
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
	photos, _ := ListFiles(dir)
	// Without GC, unclosed files stay open, so a leak shows in the fd count.
	defer debug.SetGCPercent(debug.SetGCPercent(-1))

	for name, write := range archivers {
		before := openFDs(t)
		if err := write(io.Discard, dir, photos); err != nil {
			t.Fatal(err)
		}
		if err := write(&failingWriter{n: 2000}, dir, photos); err == nil {
			t.Fatalf("%s: expected the write error to be returned", name)
		}
		if after := openFDs(t); after > before {
			t.Errorf("%s leaked %d file descriptors", name, after-before)
		}
	}
}

// BenchmarkArchive writes 20 photos of 3 MB to io.Discard, without HTTP.
func BenchmarkArchive(b *testing.B) {
	dir := b.TempDir()
	content := bytes.Repeat([]byte{0xff, 0xd8, 0x42}, 1<<20)
	for i := range 20 {
		os.WriteFile(filepath.Join(dir, fmt.Sprintf("%02d.jpg", i)), content, 0o600)
	}
	photos, _ := ListFiles(dir)
	for name, write := range archivers {
		b.Run(name, func(b *testing.B) {
			b.SetBytes(int64(20 * len(content)))
			for b.Loop() {
				if err := write(io.Discard, dir, photos); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
	b.Run("size/zip/44640", func(b *testing.B) {
		month := make([]File, 44640)
		for i := range month {
			month[i] = File{Name: fmt.Sprintf("%06d.jpg", i), Bytes: 3 << 20}
		}
		for b.Loop() {
			ZipSize(month)
		}
	})
}
