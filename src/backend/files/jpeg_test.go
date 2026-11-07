package files

import (
	"bytes"
	"image"
	"image/jpeg"
	"strings"
	"testing"
)

func encode(t *testing.T, w, h, quality int) []byte {
	t.Helper()
	img := image.NewGray(image.Rect(0, 0, w, h))
	for i := range img.Pix {
		img.Pix[i] = uint8(i * 7)
	}
	var b bytes.Buffer
	if err := jpeg.Encode(&b, img, &jpeg.Options{Quality: quality}); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func TestReadJPEGInfo(t *testing.T) {
	for _, q := range []int{5, 10, 25, 50, 75, 85, 90, 95, 100} {
		info, err := ReadJPEGInfo(bytes.NewReader(encode(t, 64, 48, q)))
		if err != nil {
			t.Fatalf("quality %d: %v", q, err)
		}
		if info.Width != 64 || info.Height != 48 || info.Quality < q-1 || info.Quality > q+1 {
			t.Errorf("quality %d: got %+v", q, info)
		}
	}
}

func TestReadJPEGInfoRejectsOtherFiles(t *testing.T) {
	for _, in := range []string{"", "jpeg\n", "\xff\xd8\xff\xd9", "\xff\xd8\xff\xdb\x00"} {
		if _, err := ReadJPEGInfo(strings.NewReader(in)); err == nil {
			t.Errorf("%q: no error", in)
		}
	}
}
