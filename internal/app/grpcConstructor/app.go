package grpcConstructor

import (
	"context"
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"imagestorage/internal/app/middleware"
	storagegrpc "imagestorage/internal/grpc/serverStorage"
)

type App struct {
	log        *logrus.Logger
	gRPCserver *grpc.Server
	port       int
}
type ImageSaver interface {
	DiskSave(ctx context.Context, imageName string, imageData []byte) error
	DeleteFile(log *logrus.Logger, imageName string, success bool)
}

func NewApp(log *logrus.Logger, port int, storage storagegrpc.Storage, diskSaver ImageSaver) *App {
	//TODO: в конфиг
	uploadDownloadSemaphore := middleware.NewSemaphore(10) // Upload/Download
	listFilesSemaphore := middleware.NewSemaphore(100)     // ListFiles

	// Создаем опции для gRPC сервера
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(middleware.UnaryInterceptor(listFilesSemaphore, log)),        // Для ListFiles
		grpc.StreamInterceptor(middleware.StreamInterceptor(uploadDownloadSemaphore, log)), // Для Upload/Download
	}

	// Создаем gRPC сервер с middleware
	grpcServer := grpc.NewServer(opts...)

	storagegrpc.RegisterServer(grpcServer, log, storage, diskSaver)

	return &App{
		log:        log,
		gRPCserver: grpcServer,
		port:       port,
	}
}

func (a *App) Start() error {
	op := "internal/app/grpcConstructor.App.Start"

	log := a.log.WithField("op", op)
	log.WithField("port", a.port).Info("Starting GRPC server")

	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := a.gRPCserver.Serve(listen); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
func (a *App) Stop() {
	const op = "internal/app/grpcConstructor.App.Stop"

	a.log.Info("stopping gRPC server, port: ", a.port, op)
	a.gRPCserver.GracefulStop()
}
