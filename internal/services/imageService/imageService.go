package imageService

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
)

type ImageService struct {
	log      *logrus.Logger
	saveDir  string
	fileLock sync.Map
}

type ImageSaver interface {
	DiskSave(ctx context.Context, imageName string, imageData []byte) error
}

func NewImageService(log *logrus.Logger, path string) *ImageService {
	return &ImageService{
		log:      log,
		saveDir:  path,
		fileLock: sync.Map{},
	}
}
func (s *ImageService) getFileLock(fileName string) *sync.Mutex {
	lock, _ := s.fileLock.LoadOrStore(fileName, &sync.Mutex{})
	return lock.(*sync.Mutex)
}
func (s *ImageService) DiskSave(ctx context.Context, imageName string, imageData []byte) error {
	op := "internal.service.ImageService.DiskSave"
	fileLock := s.getFileLock(imageName)
	fileLock.Lock()
	defer fileLock.Unlock()

	s.log.Info("Saving file on disk.... full path: ", filepath.Join(s.saveDir, imageName))

	file, err := os.OpenFile(filepath.Join(s.saveDir, imageName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	//s.log.Info(filepath.Join(s.saveDir, imageName))
	if err != nil {
		s.log.Errorf("Failed to open file: %v %s", err, op)
		return err
	}

	defer file.Close()

	_, err = file.Write(imageData)
	if err != nil {
		s.log.Errorf("Failed to write to file: %v %s", err, op)
		return err
	}

	//s.log.Infof("Successfully saved chunk to file: %s", imageName)

	return nil
}

func (s *ImageService) DeleteFile(log *logrus.Logger, imageName string, success bool) {
	fileLock := s.getFileLock(imageName)
	fileLock.Lock()
	defer fileLock.Unlock()

	filePath := filepath.Join(s.saveDir, imageName)

	if !success {
		// Если операция не завершилась успешно, удаляем файл
		if err := os.Remove(filePath); err != nil {
			log.Errorf("failed to delete file %s: %v", filePath, err)
		} else {
			log.Infof("File %s deleted due to upload failure", filePath)
		}
	}
}
