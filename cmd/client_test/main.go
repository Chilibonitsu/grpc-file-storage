package main

import (
	"context"
	"fmt"
	"imagestorage/internal/grpc/client"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
)

// unused
func main() {
	// Number of concurrent clients
	numClients := 15

	var wg sync.WaitGroup
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()

			// Connect to server
			conn, err := grpc.NewClient("localhost:57030", grpc.WithInsecure())
			if err != nil {
				log.Printf("Client %d failed to connect: %v", clientID, err)
				return
			}
			if conn != nil {
				defer conn.Close()
			}

			grpcClient := client.NewGrpcClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()

			// Upload
			filePath := fmt.Sprintf("test_file_%d.jpg", clientID)
			log.Printf("Client %d started uploading file: %s", clientID, filePath)
			err = grpcClient.UploadFile(ctx, filePath)
			if err != nil {
				log.Printf("Client %d upload error: %v", clientID, err)
			} else {
				log.Printf("Client %d uploaded file successfully", clientID)
			}

			// Download
			fileName := fmt.Sprintf("test_file_%d.jpg", clientID)
			outputPath := fmt.Sprintf("downloaded_%d.jpg", clientID)
			log.Printf("Client %d started downloading file: %s", clientID, fileName)
			err = grpcClient.DownloadFile(ctx, fileName, outputPath)
			if err != nil {
				log.Printf("Client %d download error: %v", clientID, err)
			} else {
				log.Printf("Client %d downloaded file successfully", clientID)
			}
		}(i)
	}

	wg.Wait()
	log.Println("All clients finished")
}
