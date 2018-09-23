package grpc_chunkio

import (
	"fmt"
	"io"
)

type ChunkReadStream interface {
	Recv() (*Chunk, error)
}

type chunkReader struct {
	r  ChunkReadStream
	pr *io.PipeReader
	pw *io.PipeWriter
}

func NewReader(readStream ChunkReadStream) io.ReadCloser {
	return &chunkReader{r: readStream}
}

func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pr == nil {
		cr.pr, cr.pw = io.Pipe()
		go func() {
			for {
				chunk, err := cr.r.Recv()
				if err != nil {
					if err == io.EOF {
						cr.pw.Close()
						return
					}
					cr.pw.CloseWithError(err)
					return
				}
				if chunk.Error != "" {
					cr.pw.CloseWithError(fmt.Errorf(chunk.Error))
				}
				cr.pw.Write(chunk.Data)
			}
		}()
	}
	return cr.pr.Read(p)
}

func (cr *chunkReader) Close() error {
	cr.pr.Close()
	if closer, ok := cr.r.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
