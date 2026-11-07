package files

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"math"
)

// JPEGInfo is what a photo's header says about its size on disk.
type JPEGInfo struct {
	Width, Height int
	// Quality is estimated from the luminance quantization table, on the
	// 1..100 scale of libjpeg (rpicam-still) and Go's image/jpeg.
	Quality int
}

var errNotJPEG = errors.New("not a JPEG file")

// ReadJPEGInfo reads the header of a JPEG file up to the image data.
func ReadJPEGInfo(r io.Reader) (JPEGInfo, error) {
	br := bufio.NewReader(r)
	var soi [2]byte
	if _, err := io.ReadFull(br, soi[:]); err != nil || soi != [2]byte{0xff, 0xd8} {
		return JPEGInfo{}, errNotJPEG
	}
	var info JPEGInfo
	for info.Width == 0 || info.Quality == 0 {
		marker, payload, err := nextSegment(br)
		if err != nil {
			return JPEGInfo{}, err
		}
		switch {
		case marker == 0xdb:
			if q := lumaQuality(payload); q > 0 {
				info.Quality = q
			}
		case marker >= 0xc0 && marker <= 0xcf && marker != 0xc4 && marker != 0xc8 && marker != 0xcc:
			if len(payload) < 5 {
				return JPEGInfo{}, errNotJPEG
			}
			info.Height = int(binary.BigEndian.Uint16(payload[1:]))
			info.Width = int(binary.BigEndian.Uint16(payload[3:]))
		case marker == 0xda || marker == 0xd9: // image data or end before both were found
			return JPEGInfo{}, errNotJPEG
		}
	}
	return info, nil
}

func nextSegment(br *bufio.Reader) (marker byte, payload []byte, err error) {
	b, err := br.ReadByte()
	for err == nil && b == 0xff { // fill bytes
		marker = b
		b, err = br.ReadByte()
	}
	if err != nil || marker != 0xff {
		return 0, nil, errNotJPEG
	}
	marker = b
	if marker == 0xd9 { // end of image: no length
		return marker, nil, nil
	}
	var n [2]byte
	if _, err := io.ReadFull(br, n[:]); err != nil {
		return 0, nil, errNotJPEG
	}
	size := int(binary.BigEndian.Uint16(n[:]))
	if size < 2 {
		return 0, nil, errNotJPEG
	}
	payload = make([]byte, size-2)
	if _, err := io.ReadFull(br, payload); err != nil {
		return 0, nil, errNotJPEG
	}
	return marker, payload, nil
}

// lumaBase is the luminance table of the JPEG standard (Annex K) in zigzag
// order, which libjpeg scales by quality.
var lumaBase = [64]int{
	16, 11, 12, 14, 12, 10, 16, 14, 13, 14, 18, 17, 16, 19, 24, 40,
	26, 24, 22, 22, 24, 49, 35, 37, 29, 40, 58, 51, 61, 60, 57, 51,
	56, 55, 64, 72, 92, 78, 64, 68, 87, 69, 55, 56, 80, 109, 81, 87,
	95, 98, 103, 104, 103, 62, 77, 113, 121, 112, 100, 120, 92, 101, 103, 99,
}

// lumaQuality estimates the quality from table 0 of a DQT segment, or
// returns 0 if the segment has no table 0. libjpeg scales the base table by
// 5000/quality (below 50) or 200-2*quality percent, and clamps each entry to
// 1..255; clamped entries are left out of the estimate.
func lumaQuality(dqt []byte) int {
	for len(dqt) > 0 {
		precision, id := dqt[0]>>4, dqt[0]&0x0f
		size := 64
		if precision == 1 {
			size = 128
		}
		if len(dqt) < 1+size {
			return 0
		}
		table := dqt[1 : 1+size]
		dqt = dqt[1+size:]
		if id != 0 {
			continue
		}
		var sum, base, ones int
		for i := range 64 {
			v := int(table[i])
			if precision == 1 {
				v = int(binary.BigEndian.Uint16(table[2*i:]))
			}
			switch {
			case v <= 1:
				ones++
			case v < 255:
				sum += v
				base += lumaBase[i]
			}
		}
		if base == 0 {
			if ones == 64 {
				return 100
			}
			return 1
		}
		scale := float64(sum) * 100 / float64(base)
		q := (200 - scale) / 2
		if scale > 100 {
			q = 5000 / scale
		}
		return int(min(100, max(1, math.Round(q))))
	}
	return 0
}
