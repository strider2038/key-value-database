package network

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"time"
)

type TCPServer struct {
	address        string
	semaphore      *Semaphore
	maxMessageSize int
	idleTimeout    time.Duration
	onStartup      func()
	logger         *slog.Logger
}

func NewTCPServer(
	address string,
	maxConnections int,
	maxMessageSize int,
	idleTimeout time.Duration,
	onStartup func(),
	logger *slog.Logger,
) *TCPServer {
	if onStartup == nil {
		onStartup = func() {}
	}

	return &TCPServer{
		address:        address,
		semaphore:      NewSemaphore(maxConnections),
		maxMessageSize: maxMessageSize,
		idleTimeout:    idleTimeout,
		onStartup:      onStartup,
		logger:         logger,
	}
}

func (s *TCPServer) Serve(ctx context.Context, handler Handler) error {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return fmt.Errorf("listen TCP: %w", err)
	}

	s.onStartup()

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()

		s.logger.Info("server started and ready to handle connections")

		for {
			connection, err := listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return
				}

				s.logger.Error("accept connection", "error", err)

				continue
			}

			wg.Add(1)
			s.semaphore.Acquire()
			go func(connection net.Conn) {
				defer func() {
					wg.Done()
					s.semaphore.Release()
				}()

				s.handleConnection(ctx, connection, handler)
			}(connection)
		}
	}()

	go func() {
		defer wg.Done()
		<-ctx.Done()
		if err := listener.Close(); err != nil {
			s.logger.Warn("close listener", "error", err)
		}
	}()

	wg.Wait()

	s.logger.Info("server shutdown")

	return nil
}

func (s *TCPServer) handleConnection(ctx context.Context, connection net.Conn, handler Handler) {
	request := make([]byte, s.maxMessageSize)

	for {
		deadline := time.Now().Add(s.idleTimeout)
		s.logger.Debug("set deadline", "deadline", deadline)
		if err := connection.SetDeadline(deadline); err != nil {
			s.logger.Warn("set connection deadline", "error", err)

			break
		}

		count, err := connection.Read(request)
		if err != nil {
			if err != io.EOF {
				s.logger.Warn("read from connection", "error", err)
			}

			break
		}

		response := handler.Handle(ctx, request[:count])
		if _, err := connection.Write(response); err != nil {
			s.logger.Warn("write to connection", "error", err)

			break
		}
	}

	if err := connection.Close(); err != nil {
		s.logger.Warn("close connection", "error", err)
	}
}
