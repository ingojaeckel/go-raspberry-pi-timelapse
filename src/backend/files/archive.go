package files

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"sync"
	"time"
	"unicode/utf8"
)

// ErrChanged means a photo was replaced or deleted while it was being archived.
var ErrChanged = errors.New("photo changed while archiving")

// Both archive formats store the photos uncompressed (JPEGs do not compress)
// and know their exact size before writing, so the HTTP handlers can send a
// Content-Length: browsers show progress and the time left, and a cut
// connection is detected as an incomplete download.

const tarBlock = 512

// TarSize returns the number of bytes WriteTar writes for photos.
func TarSize(photos []File) (int64, error) {
	size := int64(2 * tarBlock) // end-of-archive marker
	for _, p := range photos {
		hdr, err := tarHeader(p, 0o644)
		if err != nil {
			return 0, err
		}
		size += int64(len(hdr)) + padded(p.Bytes)
	}
	return size, nil
}

// WriteTar streams a tar archive of photos, which are in dir, to w. It
// produces the same bytes as archive/tar, but writes the file contents to w
// directly: when w is an HTTP response, net/http then sends each photo with
// sendfile(2) instead of copying it through user space.
func WriteTar(w io.Writer, dir string, photos []File) error {
	err := eachPhoto(dir, photos, func(p File, f *os.File, perm os.FileMode) error {
		hdr, err := tarHeader(p, perm)
		if err != nil {
			return err
		}
		if _, err := w.Write(hdr); err != nil {
			return err
		}
		if err := copyFile(w, f, p.Bytes); err != nil {
			return err
		}
		return writeZeros(w, padded(p.Bytes)-p.Bytes)
	})
	if err != nil {
		return err
	}
	return writeZeros(w, 2*tarBlock)
}

// tarHeader returns the header blocks archive/tar writes for p: one block,
// plus PAX records for names or times USTAR cannot hold.
func tarHeader(p File, perm os.FileMode) ([]byte, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	err := tw.WriteHeader(&tar.Header{
		Typeflag: tar.TypeReg,
		Name:     p.Name,
		Size:     p.Bytes,
		Mode:     int64(perm.Perm()),
		ModTime:  time.Unix(p.ModTimeEpoch, 0),
	})
	return buf.Bytes(), err // tw is not closed: the content follows separately
}

// padded rounds n up to whole tar blocks.
func padded(n int64) int64 {
	return (n + tarBlock - 1) / tarBlock * tarBlock
}

var zeros [2 * tarBlock]byte

func writeZeros(w io.Writer, n int64) error {
	_, err := w.Write(zeros[:n])
	return err
}

// ZipSize returns the number of bytes WriteZip writes for photos. It runs
// the zip writer without reading any photo; only the CRCs differ.
func ZipSize(photos []File) (int64, error) {
	var n countingWriter
	err := writeZip(&n, photos, func(p File, entry io.Writer) (uint32, error) {
		return 0, writeLength(entry, p.Bytes)
	})
	return int64(n), err
}

// WriteZip streams a zip archive of photos, which are in dir, to w. The
// photos are stored, and each is read once: the CRC that zip needs is
// computed while copying and written after the content, in a data descriptor.
func WriteZip(w io.Writer, dir string, photos []File) error {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	defer root.Close()
	return writeZip(w, photos, func(p File, entry io.Writer) (uint32, error) {
		var sum uint32
		err := openPhoto(root, p, func(f *os.File, _ os.FileMode) error {
			var err error
			sum, err = copyWithCRC(entry, f, p.Bytes)
			return err
		})
		return sum, err
	})
}

// writeZip writes the zip structure for photos to w; content writes each
// photo's bytes to its entry and returns their CRC-32.
func writeZip(w io.Writer, photos []File, content func(p File, entry io.Writer) (uint32, error)) error {
	zw := zip.NewWriter(w)
	for _, p := range photos {
		fh := zipHeader(p)
		entry, err := zw.CreateRaw(fh)
		if err != nil {
			return err
		}
		// zw keeps fh and writes its CRC32 into the data descriptor when
		// the entry ends, at the next CreateRaw or Close.
		if fh.CRC32, err = content(p, entry); err != nil {
			return fmt.Errorf("%s: %w", p.Name, err)
		}
	}
	return zw.Close()
}

// zipHeader describes p the way zip.Writer.CreateHeader would for a stored
// file, except that the CRC follows the content in a data descriptor.
func zipHeader(p File) *zip.FileHeader {
	// Local time like zip.Writer.CreateHeader: tools that ignore the
	// extended timestamp show the MS-DOS time as local time.
	modified := time.Unix(p.ModTimeEpoch, 0)
	fh := &zip.FileHeader{
		Name:               p.Name,
		Method:             zip.Store,
		Flags:              0x8, // sizes and CRC in a data descriptor
		CreatorVersion:     20,
		ReaderVersion:      20,
		Modified:           modified,
		CompressedSize64:   uint64(p.Bytes),
		UncompressedSize64: uint64(p.Bytes),
		Extra:              extendedTimestamp(modified),
	}
	fh.ModifiedDate, fh.ModifiedTime = msDosTime(modified)
	if !isASCII(p.Name) && utf8.ValidString(p.Name) {
		fh.Flags |= 0x800 // UTF-8 name
	}
	return fh
}

// extendedTimestamp is the "UT" extra field with the modification time in
// Unix seconds, which unzip tools prefer over the 2-second MS-DOS time.
func extendedTimestamp(t time.Time) []byte {
	b := make([]byte, 9)
	binary.LittleEndian.PutUint16(b, 0x5455)
	binary.LittleEndian.PutUint16(b[2:], 5)
	b[4] = 1 // modification time present
	binary.LittleEndian.PutUint32(b[5:], uint32(t.Unix()))
	return b
}

// msDosTime encodes t with two-second resolution, like archive/zip; MS-DOS
// times cannot be before 1980.
func msDosTime(t time.Time) (date, clock uint16) {
	if t.Year() < 1980 {
		t = time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	date = uint16(t.Day() + int(t.Month())<<5 + (t.Year()-1980)<<9)
	clock = uint16(t.Second()/2 + t.Minute()<<5 + t.Hour()<<11)
	return date, clock
}

func isASCII(s string) bool {
	for i := range len(s) {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}

// eachPhoto opens each photo inside dir in turn, checks that it is still the
// file that was listed, and closes it before moving on to the next one.
func eachPhoto(dir string, photos []File, fn func(p File, f *os.File, perm os.FileMode) error) error {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	defer root.Close()
	for _, p := range photos {
		if err := openPhoto(root, p, func(f *os.File, perm os.FileMode) error { return fn(p, f, perm) }); err != nil {
			return fmt.Errorf("%s: %w", p.Name, err)
		}
	}
	return nil
}

func openPhoto(root *os.Root, p File, fn func(f *os.File, perm os.FileMode) error) error {
	f, err := root.Open(p.Name)
	if err != nil {
		return err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return ErrNotAFile
	}
	if info.Size() != p.Bytes {
		return fmt.Errorf("%w: now %d bytes, listed with %d", ErrChanged, info.Size(), p.Bytes)
	}
	return fn(f, info.Mode().Perm())
}

// copyFile copies exactly n bytes of f to w. The *io.LimitedReader of an
// *os.File is what net/http and net hand to sendfile(2).
func copyFile(w io.Writer, f *os.File, n int64) error {
	copied, err := io.Copy(w, &io.LimitedReader{R: f, N: n})
	if err == nil && copied != n {
		err = fmt.Errorf("%w: ended after %d of %d bytes", ErrChanged, copied, n)
	}
	return err
}

// copyBufferSize matches the Linux default read-ahead, so each read is
// served from one read-ahead window of the SD card.
const copyBufferSize = 128 << 10

var copyBuffers = sync.Pool{New: func() any { b := make([]byte, copyBufferSize); return &b }}

// copyWithCRC copies exactly n bytes of r to w and returns their CRC-32.
func copyWithCRC(w io.Writer, r io.Reader, n int64) (uint32, error) {
	bp := copyBuffers.Get().(*[]byte)
	defer copyBuffers.Put(bp)
	crc := crc32.NewIEEE()
	copied, err := io.CopyBuffer(io.MultiWriter(w, crc), io.LimitReader(r, n), *bp)
	if err == nil && copied != n {
		err = fmt.Errorf("%w: ended after %d of %d bytes", ErrChanged, copied, n)
	}
	return crc.Sum32(), err
}

// writeLength writes n bytes of nothing in particular to w, for ZipSize.
func writeLength(w io.Writer, n int64) error {
	bp := copyBuffers.Get().(*[]byte)
	defer copyBuffers.Put(bp)
	for n > 0 {
		chunk := min(n, int64(len(*bp)))
		if _, err := w.Write((*bp)[:chunk]); err != nil {
			return err
		}
		n -= chunk
	}
	return nil
}

type countingWriter int64

func (c *countingWriter) Write(p []byte) (int, error) {
	*c += countingWriter(len(p))
	return len(p), nil
}
