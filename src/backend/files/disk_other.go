//go:build !unix

package files

import "math"

// FreeBytes is not implemented on this platform; it reports unlimited space
// so that Reclaim never deletes photos.
func FreeBytes(string) (uint64, error) {
	return math.MaxUint64, nil
}
