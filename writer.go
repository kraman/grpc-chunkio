package grpc_chunkio

import "io"

// ChunkWriteStream represents a gRPC stream which can send Chunks
type ChunkWriteStream interface {
	Send(*Chunk) error
}

type chunkWriter struct {
	w          ChunkWriteStream
	packetSize int
	c          Chunk
}

// NewWriter returns a WriteCloser which will send data over the gRPC Stream
func NewWriter(responseStream ChunkWriteStream, packetSize int) io.WriteCloser {
	return &chunkWriter{w: responseStream, packetSize: packetSize}
}

func (cw *chunkWriter) Write(p []byte) (n int, err error) {
	l := len(p)
	if l > cw.packetSize {
		if err = cw.w.Send(&Chunk{Data: p[0:cw.packetSize]}); err != nil {
			return 0, err
		}
		return cw.packetSize, nil
	}
	if err = cw.w.Send(&Chunk{Data: p}); err != nil {
		return 0, err
	}
	return l, nil
}

func (cw *chunkWriter) Close() error {
	if closer, ok := cw.w.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
