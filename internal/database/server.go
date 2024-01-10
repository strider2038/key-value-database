package database

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type Service interface {
	Serve(ctx context.Context) error
}

type Server struct {
	services []Service
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) AddService(service Service) {
	s.services = append(s.services, service)
}

func (s *Server) Serve(ctx context.Context) error {
	if len(s.services) == 0 {
		return fmt.Errorf("empty services")
	}

	serveContext, cancel := context.WithCancel(ctx)
	defer cancel()

	mu := sync.Mutex{}
	var errs []error

	waiter := sync.WaitGroup{}
	waiter.Add(len(s.services))
	for _, service := range s.services {
		go func(s Service) {
			defer waiter.Done()
			if err := s.Serve(serveContext); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				cancel()
			}
		}(service)
	}
	waiter.Wait()

	return errors.Join(errs...)
}
