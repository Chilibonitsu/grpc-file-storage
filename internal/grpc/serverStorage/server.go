package serverStorage

import (
	"context"
	"imagestorage/internal/storage/sqlite"
	"imagestorage/internal/utils"
	"io"
	"os"
	"path/filepath"
	"time"

	pb "github.com/Chilibonitsu/protos/gen/go/imageStorage"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Storage interface {
	SaveImage(imageName string, size int, mimeType string, checksum string, createdAt time.Time) error
	ListFiles() ([]sqlite.FileInfo, error)
	FindFileByName(fileName string) (string, error)
}

type ImageSaver interface {
	DiskSave(ctx context.Context, imageName string, imageData []byte) error
	DeleteFile(log *logrus.Logger, imageName string, success bool)
}

type serverAPI struct {
	pb.UnimplementedGuploadServiceServer
	log       *logrus.Logger
	storage   Storage
	diskSaver ImageSaver
}

func RegisterServer(gRPC *grpc.Server, log *logrus.Logger, storage Storage, diskSaver ImageSaver) {
	server := &serverAPI{storage: storage, log: log, diskSaver: diskSaver}
	pb.RegisterGuploadServiceServer(gRPC, server)
}

// TODO conf
const maxImageSize = 1024 * 1024 * 10 // 10MB

func (s *serverAPI) Upload(stream pb.GuploadService_UploadServer) error {
	op := "internal.grpc.ServerStorage.Upload"
	var fileName string

	ctx := stream.Context()
	//сделать таймаут в interceptor
	// timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	// defer cancel()

	req, err := stream.Recv()
	if err != nil {
		s.log.Errorf("failed to receive image info: %v", err)
		return status.Errorf(codes.Unknown, "failed to receive image info: %v", err)
	}

	if fileInfo := req.GetData().(*pb.UploadFileRequest_FileInfo); fileInfo != nil {
		fileName = fileInfo.FileInfo.GetFileName()
	} else {
		return status.Errorf(codes.InvalidArgument, "missing file info")
	}

	s.log.Info("Received file name: ", fileName, op)
	ok := utils.CheckFileName(fileName)
	if !ok {
		return status.Errorf(codes.InvalidArgument, "invalid file name")
	}

	fileExtension := utils.GetExt(fileName)
	s.log.Infof("Recived file name: %s, extension %s", fileName, fileExtension)

	//Что делать с файлами с одинаковым названием?
	//Если в базе есть файл с таким названием, то мы его не загружаем
	//

	findFileName, err := s.storage.FindFileByName(fileName)
	if err != nil {
		return status.Errorf(codes.Internal, "internal error: %v", err)
	}

	if len(findFileName) > 0 {
		return status.Errorf(codes.AlreadyExists, "file already exists: %s", findFileName)
	}
	//

	var imageSize int
	imageSize = 0

	var success bool //Если запрос не дойдет до конца, то удалим файл(если мы получим не все данные, ошибка в базе etc)
	success = false

	defer func() {
		defer s.diskSaver.DeleteFile(s.log, fileName, success)
	}()

	for {
		s.log.Info("Waiting for file data...")
		req, err := stream.Recv()

		if err == io.EOF {
			break
		}
		if err != nil {
			s.log.Error(op, err)
			return status.Errorf(codes.Internal, "failed to receive file data: %v", err)
		}

		chunk := req.GetContent()
		size := len(chunk)

		if size > maxImageSize {
			return status.Errorf(codes.InvalidArgument, "file size exceeds the maximum allowed size")
		}

		imageSize += len(chunk)

		err = s.diskSaver.DiskSave(ctx, fileName, chunk)
		if err != nil {
			s.log.Errorf("failed to save image: %v", err)
			return status.Errorf(codes.Internal, "failed to save image: %v", err)
		}

	}

	checksumm, err := utils.CalculateChecksum(fileName)
	if err != nil {
		s.log.Errorf("failed to calculate checksum: %v", err)
	}

	err = s.storage.SaveImage(fileName, imageSize, fileExtension, checksumm, time.Now())

	if err != nil {
		s.log.Errorf("failed to save image info: %v", err)
		return status.Errorf(codes.Internal, "failed to save image info: %v", err)
	}

	response := &pb.UploadResponse{
		Message: "File uploaded successfully",
		Code:    pb.UploadStatusCode_Ok,
	}

	err = stream.SendAndClose(response)
	if err != nil {
		s.log.Errorf("failed to send response: %v", err)
		return status.Errorf(codes.Internal, "failed to send response: %v", err)
	}

	s.log.Info("File uploaded successfully ", fileName, "size KB: ", imageSize)

	success = true
	return err
}

func (s *serverAPI) Download(req *pb.DownloadRequest, stream pb.GuploadService_DownloadServer) error {
	//ctx := stream.Context()

	fileName := req.GetFileName()
	if fileName == "" {
		return status.Errorf(codes.InvalidArgument, "file name is required")
	}

	fileName = filepath.Join(os.Getenv("PATH_TO_SAVED"), fileName)
	file, err := os.Open(fileName)

	if err != nil {
		if os.IsNotExist(err) {
			return status.Errorf(codes.NotFound, "file not found: %s", fileName)
		}
		return status.Errorf(codes.Internal, "failed to open file: %v", err)
	}
	defer file.Close()

	//TODO: брать из конфига
	buffer := make([]byte, 1024*64)
	for {
		n, err := file.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			return status.Errorf(codes.Internal, "failed to read file: %v", err)
		}

		chunk := &pb.DownloadResponse{
			Content: buffer[:n],
		}

		if err := stream.Send(chunk); err != nil {
			return status.Errorf(codes.Internal, "failed to send chunk: %v", err)
		}
	}

	return nil
}

func (s *serverAPI) ListFiles(ctx context.Context, req *pb.ListFilesRequest) (*pb.ListFilesResponse, error) {
	files, err := s.storage.ListFiles()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list files: %v", err)
	}

	var fileInfos []*pb.FileInfo
	for _, dbFileInfo := range files {
		fileInfo := &pb.FileInfo{
			FileName:  dbFileInfo.FileName,
			CreatedAt: dbFileInfo.CreatedAt.String(),
			UpdatedAt: dbFileInfo.CreatedAt.String(),
		}
		fileInfos = append(fileInfos, fileInfo)
	}

	return &pb.ListFilesResponse{Files: fileInfos}, nil
}
