package client

import (
	"context"
	"fmt"

	pb "github.com/pdf-rag-system/backend/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type DocReaderClient struct {
	conn   *grpc.ClientConn
	client pb.DocReaderClient
}

func NewDocReaderClient(host, port string) (*DocReaderClient, error) {
	addr := fmt.Sprintf("%s:%s", host, port)

	// Set max message size to 100MB (to handle large PDFs)
	maxMsgSize := 100 * 1024 * 1024 // 100 MB
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxMsgSize),
			grpc.MaxCallSendMsgSize(maxMsgSize),
		),
	}

	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to docreader: %w", err)
	}

	return &DocReaderClient{
		conn:   conn,
		client: pb.NewDocReaderClient(conn),
	}, nil
}

func (c *DocReaderClient) ParsePDF(ctx context.Context, fileContent []byte, filename string, chunkSize, chunkOverlap int32) (*pb.ParseResponse, error) {
	req := &pb.ParseRequest{
		FileContent: fileContent,
		Filename:    filename,
		ChunkConfig: &pb.ChunkConfig{
			ChunkSize:    chunkSize,
			ChunkOverlap: chunkOverlap,
		},
	}

	return c.client.ParsePDF(ctx, req)
}

func (c *DocReaderClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
