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

type TCPHandler = func(context.Context, []byte) []byte

type TCPServer struct {
	address        string
	semaphore      *Semaphore
	maxMessageSize int
	idleTimeout    time.Duration
	logger         *slog.Logger
}

func NewTCPServer(
	address string,
	maxConnections int,
	maxMessageSize int,
	idleTimeout time.Duration,
	logger *slog.Logger,
) *TCPServer {
	return &TCPServer{
		address:        address,
		semaphore:      NewSemaphore(maxConnections),
		maxMessageSize: maxMessageSize,
		idleTimeout:    idleTimeout,
		logger:         logger,
	}
}

func (s *TCPServer) Serve(ctx context.Context, handler TCPHandler) error {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return fmt.Errorf("listen TCP: %w", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()

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

	return nil
}

func (s *TCPServer) handleConnection(ctx context.Context, connection net.Conn, handler TCPHandler) {
	request := make([]byte, s.maxMessageSize)

	for {
		if err := connection.SetDeadline(time.Now().Add(s.idleTimeout)); err != nil {
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

		response := handler(ctx, request[:count])
		if _, err := connection.Write(response); err != nil {
			s.logger.Warn("write to connection", "error", err)

			break
		}
	}

	if err := connection.Close(); err != nil {
		s.logger.Warn("close connection", "error", err)
	}
}
