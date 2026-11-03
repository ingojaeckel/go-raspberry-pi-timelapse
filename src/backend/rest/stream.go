package rest

import (
	"io"
	"math"
	"net/http"
	"time"
)

// stallTimeout is how long a download may make no progress before the
// connection is dropped. Without it a phone that leaves the WiFi keeps its
// download (a goroutine, a socket and an open photo) alive until TCP gives
// up retransmitting, which takes up to half an hour. A variable for tests.
var stallTimeout = time.Minute

// sendChunk is how much is handed to sendfile(2) per write deadline, so
// that a client that reads a few KB/s is never cut off: the deadline is
// at least 3/4 stallTimeout ahead when a chunk starts.
const sendChunk = 256 << 10

// deadlineWriter extends the connection's write deadline while a response
// makes progress. It keeps net/http's sendfile path: ReadFrom passes the
// *io.LimitedReader of an *os.File through, only in smaller pieces.
type deadlineWriter struct {
	http.ResponseWriter
	rc       *http.ResponseController
	extended time.Time
}

func newDeadlineWriter(w http.ResponseWriter) *deadlineWriter {
	d := &deadlineWriter{ResponseWriter: w, rc: http.NewResponseController(w)}
	d.extend()
	return d
}

// extend moves the deadline to stallTimeout from now, at most every
// stallTimeout/4 to keep timer updates cheap.
func (d *deadlineWriter) extend() {
	now := time.Now()
	if now.Sub(d.extended) < stallTimeout/4 {
		return
	}
	d.extended = now
	_ = d.rc.SetWriteDeadline(now.Add(stallTimeout)) // not supported by test recorders
}

func (d *deadlineWriter) Write(p []byte) (int, error) {
	d.extend()
	return d.ResponseWriter.Write(p)
}

func (d *deadlineWriter) ReadFrom(src io.Reader) (int64, error) {
	lr, ok := src.(*io.LimitedReader)
	if !ok {
		lr = &io.LimitedReader{R: src, N: math.MaxInt64}
	}
	var total int64
	for lr.N > 0 {
		d.extend()
		want := min(lr.N, sendChunk)
		n, err := io.Copy(d.ResponseWriter, &io.LimitedReader{R: lr.R, N: want})
		total += n
		lr.N -= n
		if err != nil || n < want {
			return total, err
		}
	}
	return total, nil
}

// Unwrap lets http.ResponseController reach the underlying writer.
func (d *deadlineWriter) Unwrap() http.ResponseWriter { return d.ResponseWriter }
