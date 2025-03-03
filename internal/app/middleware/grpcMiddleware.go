package middleware

import (
	"context"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Semaphore struct {
	sem chan struct{}
}

func NewSemaphore(maxConcurrentRequests int) Semaphore {
	return Semaphore{
		sem: make(chan struct{}, maxConcurrentRequests),
	}
}

func (s Semaphore) Acquire() error {
	select {
	case s.sem <- struct{}{}:
		return nil
	default:
		return status.Errorf(codes.ResourceExhausted, "too many concurrent requests")
	}
}
func (s *Semaphore) Release() {
	<-s.sem
}

func UnaryInterceptor(sem Semaphore, log *logrus.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if err := sem.Acquire(); err != nil {
			log.WithFields(logrus.Fields{
				"error": err,
			}).Warn("Too many concurrent requests, limit reached!")

			return nil, err
		}
		defer sem.Release()

		return handler(ctx, req)
	}
}

func StreamInterceptor(sem Semaphore, log *logrus.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if err := sem.Acquire(); err != nil {
			log.WithFields(logrus.Fields{
				"error": err,
			}).Warn("Too many concurrent requests, limit reached!")
			return err
		}
		defer sem.Release()
		log.Info("StreamInterceptor called")
		return handler(srv, stream)
	}
}
