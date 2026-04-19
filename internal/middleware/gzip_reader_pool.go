package middleware

import (
	"compress/gzip"
	"io"
	"sync"
)

var gzipReaderPool = sync.Pool{
	New: func() any {
		return new(gzip.Reader)
	},
}

func getGzipReader(r io.Reader) (*gzip.Reader, error) {
	gr := gzipReaderPool.Get().(*gzip.Reader)
	if err := gr.Reset(r); err != nil {
		return nil, err
	}
	return gr, nil
}

func putGzipReader(gr *gzip.Reader) {
	_ = gr.Close()
	gzipReaderPool.Put(gr)
}
