package network_test

import (
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/strider2038/key-value-database/internal/database/network"
)

type EchoServer struct {
	Listener        net.Listener
	ResponseTimeout time.Duration
}

func NewEchoServer(address string) (*EchoServer, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("listen TCP: %w", err)
	}

	return &EchoServer{Listener: listener}, nil
}

func (s *EchoServer) Run() error {
	for {
		connection, err := s.Listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}

			return fmt.Errorf("accept connection: %w", err)
		}
		if err := s.handleConnection(connection); err != nil {
			return err
		}
	}
}

func (s *EchoServer) handleConnection(connection net.Conn) error {
	defer connection.Close()

	request := make([]byte, messageSize)
	count, err := connection.Read(request)
	if err != nil {
		return fmt.Errorf("read from connection: %w", err)
	}

	time.Sleep(s.ResponseTimeout)

	response := append([]byte("echo to "), request[:count]...)
	if _, err := connection.Write(response); err != nil {
		return fmt.Errorf("write to connection: %w", err)
	}

	return nil
}

func (s *EchoServer) Close() error {
	return s.Listener.Close()
}

func TestTCPClient_Send(t *testing.T) {
	const address = ":10002"
	server, err := NewEchoServer(address)
	require.NoError(t, err)
	defer server.Close()
	go func() {
		require.NoError(t, server.Run())
	}()

	client, err := network.NewTCPClient(address, messageSize, time.Second)
	require.NoError(t, err)
	response, err := client.Send([]byte("request"))
	require.NoError(t, err)

	assert.Equal(t, "echo to request", string(response))
}

func TestTCPClient_Send_IdleTimeout(t *testing.T) {
	const address = ":10003"
	server, err := NewEchoServer(address)
	require.NoError(t, err)
	server.ResponseTimeout = 100 * time.Millisecond
	defer server.Close()
	go func() {
		require.NoError(t, server.Run())
	}()

	client, err := network.NewTCPClient(address, messageSize, 10*time.Millisecond)
	require.NoError(t, err)
	_, err = client.Send([]byte("request"))

	var timeoutErr interface{ Timeout() bool }
	require.ErrorAs(t, err, &timeoutErr)
	assert.True(t, timeoutErr.Timeout())
}
