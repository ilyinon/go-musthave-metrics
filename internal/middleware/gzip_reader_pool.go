package middleware

import (
	"compress/gzip"
	"io"
	"sync"
)

// gzipReaderPool is used to reuse gzip.Reader instances
// to reduce memory allocations.
var gzipReaderPool = sync.Pool{
	New: func() any {
		return new(gzip.Reader)
	},
}

// getGzipReader retrieves a gzip.Reader from the pool and resets it with the provided reader.
func getGzipReader(r io.Reader) (*gzip.Reader, error) {
	gr := gzipReaderPool.Get().(*gzip.Reader)
	if err := gr.Reset(r); err != nil {
		return nil, err
	}
	return gr, nil
}

// putGzipReader returns a gzip.Reader to the pool after closing it.
func putGzipReader(gr *gzip.Reader) {
	_ = gr.Close()
	gzipReaderPool.Put(gr)
}
