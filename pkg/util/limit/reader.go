package limit

import (
	"context"
	"golang.org/x/time/rate"
	"io"
)

type Reader struct {
	r       io.Reader
	limiter *rate.Limiter
}

func NewReader(r io.Reader, limiter *rate.Limiter) *Reader {
	return &Reader{
		r:       r,
		limiter: limiter,
	}
}

func (r *Reader) Read(p []byte) (n int, err error) {
	b := r.limiter.Burst()
	if b < len(p) {
		p = p[:b]
	}
	n, err = r.r.Read(p)
	if err != nil {
		return
	}
	err = r.limiter.WaitN(context.Background(), n)
	if err != nil {
		return
	}
	return
}
