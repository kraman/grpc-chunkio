package main

import (
	"bufio"
	"io"
	"net"
	"os"

	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	chunkio "github.com/kraman/grpc-chunkio"
)

// client
func downloadFile(srcPath, destPath string) error {
	f, err := os.OpenFile(destPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0660)
	if err != nil {
		return err
	}

	conn, err := grpc.Dial("localhost:30001", grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	client := NewFileTransferClient(conn)
	streamClient, err := client.Get(context.Background(), &GetRequest{Path: srcPath})
	if err != nil {
		return err
	}
	r := chunkio.NewReader(streamClient)
	_, err = io.Copy(bufio.NewWriter(f), r)
	return err
}

// server
type impl struct{}

func (i *impl) Get(req *GetRequest, respstream FileTransfer_GetServer) error {
	f, err := os.OpenFile(req.GetPath(), os.O_RDONLY, 0660)
	if err != nil {
		return respstream.Send(&chunkio.Chunk{Error: err.Error()})
	}
	writer := bufio.NewWriter(chunkio.NewWriter(respstream, 1024*64))

	if _, err = io.Copy(writer, bufio.NewReader(f)); err != nil {
		return respstream.Send(&chunkio.Chunk{Error: err.Error()})
	}
	return nil
}

func listen(ctx context.Context, l string) error {
	tcpl, err := net.Listen("tcp", l)
	if err != nil {
		return errors.WithMessage(err, "error listening on host: "+l)
	}

	s := grpc.NewServer()
	RegisterFileTransferServer(s, &impl{})

	g, _ := errgroup.WithContext(ctx)
	g.Go(func() error { return s.Serve(tcpl) })

	return g.Wait()
}
