package rest

import (
	"testing"

	"github.com/facebookgo/ensure"
)

func TestGetHumanReadableSize(t *testing.T) {
	ensure.DeepEqual(t, "0 bytes", getHumanReadableSize(0))
	ensure.DeepEqual(t, "1023 bytes", getHumanReadableSize(1023))
	ensure.DeepEqual(t, "1 KB", getHumanReadableSize(1024))
	ensure.DeepEqual(t, "10 KB", getHumanReadableSize(10*1024))
	ensure.DeepEqual(t, "1 MB", getHumanReadableSize(1024*1024))
}

// Benchmarks for performance-sensitive operations

func BenchmarkGetHumanReadableSize(b *testing.B) {
	testSizes := []int64{
		0,
		1023,
		1024,
		10 * 1024,
		1024 * 1024,
		10 * 1024 * 1024,
		100 * 1024 * 1024,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, size := range testSizes {
			_ = getHumanReadableSize(size)
		}
	}
}
