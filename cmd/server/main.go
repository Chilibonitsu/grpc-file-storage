package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"imagestorage/internal/app"
	"imagestorage/internal/config"
	"imagestorage/internal/services/imageService"
	"imagestorage/internal/storage/sqlite"

	pb "imagestorage/contracts/gen/go/imageStorage"

	"imagestorage/internal/grpc/client"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)
	log.Info(cfg)
	log.Info("Starting server...")

	imageDB, err := sqlite.New(cfg.StoragePath)

	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	diskSaver := imageService.NewImageService(log, cfg.ServerImageStorage)
	GRPCport, err := strconv.Atoi(cfg.GRPC.Port)
	if err != nil {
		log.Fatal(err)
	}
	storeImageServer := app.NewApp(log, GRPCport, imageDB, diskSaver)

	go storeImageServer.GRPCsrv.Start()

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	cl, err := grpc.NewClient(cfg.GRPC.Address+":"+cfg.GRPC.Port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error(err)
	}

	grpcClient := client.NewGrpcClient(cl)
	err = grpcClient.UploadFile(ctx, "test_file_client.jpg")
	if err != nil {
		log.Error(err)
	}

	err = grpcClient.UploadFile(ctx, "image.jpg")
	if err != nil {
		log.Error(err)
	}

	err = grpcClient.DownloadFile(ctx, "test_file_client.jpg", cfg.ClientImageStorage)
	if err != nil {
		log.Error(err)
	}
	_, err = grpcClient.ListFiles(ctx)
	if err != nil {
		log.Error(err)
	}

	<-stop

}

// conn, err := grpc.Dial("localhost:57030", grpc.WithInsecure(), grpc.WithBlock())
// if err != nil {
// 	log.Fatalf("Failed to connect to server: %v", err)
// }
// defer conn.Close()

// client := pb.NewGuploadServiceClient(conn)

// // Вызываем метод для скачивания изображения
// fileName := "SDVM6zOYSik.jpg" // Имя файла, который хотим скачать
// err = downloadImage(client, fileName)
// if err != nil {
// 	log.Fatalf("Failed to download image: %v", err)
// }

func downloadImage(client pb.GuploadServiceClient, fileName string) error {
	// Создаем запрос на скачивание файла
	req := &pb.DownloadRequest{
		FileName: fileName,
	}

	// Вызываем метод Download на сервере
	stream, err := client.Download(context.Background(), req)
	if err != nil {
		return fmt.Errorf("failed to call Download: %v", err)
	}

	// Создаем файл для сохранения изображения
	outputPath := filepath.Join(os.Getenv("PATH_TO_SAVED_CLIENT"), fileName)
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Получаем данные от сервера и записываем их в файл
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to receive chunk: %v", err)
		}

		_, err = file.Write(chunk.Content)
		if err != nil {
			return fmt.Errorf("failed to write chunk to file: %v", err)
		}
	}

	return nil
}

func setupLogger(env string) *logrus.Logger {
	log := logrus.New()

	switch env {
	case envLocal:
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:          true,
			ForceColors:            true,
			DisableQuote:           true,
			DisableLevelTruncation: true,
			PadLevelText:           true,
			QuoteEmptyFields:       true,
			DisableTimestamp:       false,
		})
	case envDev:
		log.SetFormatter(&logrus.JSONFormatter{

			DisableTimestamp:  false,
			DisableHTMLEscape: true,
			DataKey:           "data",
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
				logrus.FieldKeyFunc:  "caller",
			},
		})
	case envProd:
		log.SetLevel(logrus.WarnLevel)
	}
	return log
}
