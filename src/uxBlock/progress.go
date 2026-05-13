package uxBlock

import (
	"context"
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"github.com/zeropsio/zcli/src/units"
)

// ProgressReader wraps r and starts a goroutine that periodically
// updates h's spinner line with a running byte counter formatted as
// "<label> (X.X MiB)". The returned stop function cancels the
// updater; callers should defer it inside the WithSpinner closure to
// guarantee the goroutine exits regardless of how the upload finishes.
//
// Refresh cadence is 500ms — fast enough to feel live, slow enough
// not to flood the terminal under high-throughput uploads.
func ProgressReader(h SpinnerHandle, label string, r io.Reader) (io.Reader, func()) {
	cr := &countingReader{r: r}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		t := time.NewTicker(500 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				n := cr.n.Load()
				if n < 0 {
					n = 0
				}
				h.SetMessage(fmt.Sprintf("%s (%s)", label, units.ByteCountIEC(uint64(n))))
			}
		}
	}()
	return cr, cancel
}

type countingReader struct {
	r io.Reader
	n atomic.Int64
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n.Add(int64(n))
	return n, err
}
