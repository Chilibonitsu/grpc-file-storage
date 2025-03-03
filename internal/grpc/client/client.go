package client

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	pb "github.com/Chilibonitsu/protos/gen/go/imageStorage"

	"google.golang.org/grpc"
)

type GrpcClient struct {
	client pb.GuploadServiceClient
}

func NewGrpcClient(conn *grpc.ClientConn) *GrpcClient {
	return &GrpcClient{
		client: pb.NewGuploadServiceClient(conn),
	}
}

func (c *GrpcClient) UploadFile(ctx context.Context, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	stream, err := c.client.Upload(ctx)
	if err != nil {
		return fmt.Errorf("failed to create upload stream: %v", err)
	}

	fileInfo := &pb.UploadFileRequest{
		Data: &pb.UploadFileRequest_FileInfo{
			FileInfo: &pb.FileUploadInfo{
				FileName: filepath.Base(filePath),
			},
		},
	}
	if err := stream.Send(fileInfo); err != nil {
		return fmt.Errorf("failed to send file info: %v", err)
	}

	// Читаем и отправляем данные файла по частям
	buffer := make([]byte, 64*1024)
	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading file: %v", err)
		}

		chunk := &pb.UploadFileRequest{
			Data: &pb.UploadFileRequest_Content{
				Content: buffer[:n],
			},
		}
		if err := stream.Send(chunk); err != nil {
			return fmt.Errorf("failed to send chunk: %v", err)
		}
	}

	// Получаем ответ от сервера
	response, err := stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("failed to receive response: %v", err)
	}

	if response.Code != pb.UploadStatusCode_Ok {
		return fmt.Errorf("upload failed with code: %v", response.Code)
	}

	return nil
}

func (c *GrpcClient) DownloadFile(ctx context.Context, fileName string, outputPath string) error {
	request := &pb.DownloadRequest{
		FileName: fileName,
	}

	stream, err := c.client.Download(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to start download: %v", err)
	}

	out := filepath.Join(outputPath, fileName)

	file, err := os.Create(out)

	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer file.Close()

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error receiving chunk: %v", err)
		}

		if _, err := file.Write(chunk.Content); err != nil {
			return fmt.Errorf("failed to write chunk: %v", err)
		}
	}

	return nil
}

func (c *GrpcClient) ListFiles(ctx context.Context) ([]*pb.FileInfo, error) {
	response, err := c.client.ListFiles(ctx, &pb.ListFilesRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %v", err)
	}
	return response.Files, nil
}
