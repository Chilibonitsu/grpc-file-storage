package main

import (
	"context"
	"log"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"imagestorage/internal/app/middleware"
)

type MockGrpcClient struct {
	uploadSemaphore    middleware.Semaphore
	downloadSemaphore  middleware.Semaphore
	listFilesSemaphore middleware.Semaphore
	log                *log.Logger
}

func NewMockGrpcClient(uploadLimit, downloadLimit, listFilesLimit int) *MockGrpcClient {
	return &MockGrpcClient{
		uploadSemaphore:    middleware.NewSemaphore(uploadLimit),
		downloadSemaphore:  middleware.NewSemaphore(downloadLimit),
		listFilesSemaphore: middleware.NewSemaphore(listFilesLimit),
		log:                log.Default(),
	}
}

func (m *MockGrpcClient) UploadFile(ctx context.Context, fileName string) error {
	if err := m.uploadSemaphore.Acquire(); err != nil {
		m.log.Printf("UploadFile: too many concurrent requests: %v", err)
		return err
	}
	defer m.uploadSemaphore.Release()

	// Имитация загрузки файла
	time.Sleep(400 * time.Millisecond)
	return nil
}

func (m *MockGrpcClient) DownloadFile(ctx context.Context, fileName, storagePath string) error {
	if err := m.downloadSemaphore.Acquire(); err != nil {
		m.log.Printf("DownloadFile: too many concurrent requests: %v", err)
		return err
	}
	defer m.downloadSemaphore.Release()

	// Имитация скачивания файла
	time.Sleep(400 * time.Millisecond)
	return nil
}

func (m *MockGrpcClient) ListFiles(ctx context.Context) ([]string, error) {
	if err := m.listFilesSemaphore.Acquire(); err != nil {
		m.log.Printf("ListFiles: too many concurrent requests: %v", err)
		return nil, err
	}
	defer m.listFilesSemaphore.Release()

	// Имитация получения списка файлов
	time.Sleep(400 * time.Millisecond)
	return []string{"file1", "file2"}, nil
}

func TestConcurrentAccess(t *testing.T) {
	// Создаем мок клиента с ограничениями на количество одновременных запросов
	mockClient := NewMockGrpcClient(9, 9, 9)

	numGoroutines := 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			ctx := context.Background()

			// Имитация загрузки файла
			err := mockClient.UploadFile(ctx, "test_file.jpg")
			if err != nil {
				if status.Code(err) == codes.ResourceExhausted {
					t.Logf("Goroutine %d: UploadFile: resource exhausted (expected)", id)
				} else {
					t.Errorf("Goroutine %d: UploadFile: unexpected error: %v", id, err)
				}
			} else {
				t.Logf("Goroutine %d: UploadFile: success", id)
			}

			// Имитация скачивания файла
			err = mockClient.DownloadFile(ctx, "test_file.jpg", "/tmp")
			if err != nil {
				if status.Code(err) == codes.ResourceExhausted {
					t.Logf("Goroutine %d: DownloadFile: resource exhausted (expected)", id)
				} else {
					t.Errorf("Goroutine %d: DownloadFile: unexpected error: %v", id, err)
				}
			} else {
				t.Logf("Goroutine %d: DownloadFile: success", id)
			}

			// Имитация получения списка файлов
			files, err := mockClient.ListFiles(ctx)
			if err != nil {
				if status.Code(err) == codes.ResourceExhausted {
					t.Logf("Goroutine %d: ListFiles: resource exhausted (expected)", id)
				} else {
					t.Errorf("Goroutine %d: ListFiles: unexpected error: %v", id, err)
				}
			} else {
				t.Logf("Goroutine %d: ListFiles: success, files: %v", id, files)
			}
		}(i)
	}

	wg.Wait()
}
