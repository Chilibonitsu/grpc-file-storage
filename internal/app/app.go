package app

import (
	"github.com/sirupsen/logrus"

	grpcConstructor "imagestorage/internal/app/grpcConstructor"
	storagegrpc "imagestorage/internal/grpc/serverStorage"
)

type App struct {
	GRPCsrv *grpcConstructor.App
}

func NewApp(log *logrus.Logger, grpcPort int, storage storagegrpc.Storage, diskSaver storagegrpc.ImageSaver) *App {
	// TODO: хранилище

	//init image storage

	grpcApp := grpcConstructor.NewApp(log, grpcPort, storage, diskSaver)
	return &App{
		GRPCsrv: grpcApp,
	}
}
