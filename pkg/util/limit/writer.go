package limit

import (
	"context"
	"golang.org/x/time/rate"
	"io"
)

type Writer struct {
	w       io.Writer
	limiter *rate.Limiter
}

func NewWriter(w io.Writer, limiter *rate.Limiter) *Writer {
	return &Writer{
		w:       w,
		limiter: limiter,
	}
}

func (w *Writer) Write(p []byte) (n int, err error) {
	var nn int
	b := w.limiter.Burst()
	for {

		end := len(p)
		if end == 0 {
			break
		}
		if b < len(p) {
			end = b
		}
		err = w.limiter.WaitN(context.Background(), end)
		if err != nil {
			return
		}
		nn, err = w.w.Write(p[:end])
		n += nn
		if err != nil {
			return
		}
		p = p[end:]
	}
	return
}
