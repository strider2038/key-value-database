package network

import (
	"fmt"
	"net"
	"time"
)

type TCPClient struct {
	connection     net.Conn
	maxMessageSize int
	idleTimeout    time.Duration
}

func NewTCPClient(address string, maxMessageSize int, idleTimeout time.Duration) (*TCPClient, error) {
	connection, err := net.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("dial TCP: %w", err)
	}

	return &TCPClient{
		connection:     connection,
		maxMessageSize: maxMessageSize,
		idleTimeout:    idleTimeout,
	}, nil
}

func (c *TCPClient) Send(request []byte) ([]byte, error) {
	if err := c.connection.SetDeadline(time.Now().Add(c.idleTimeout)); err != nil {
		return nil, fmt.Errorf("set connection deadline: %w", err)
	}

	if _, err := c.connection.Write(request); err != nil {
		return nil, fmt.Errorf("write to connection: %w", err)
	}

	response := make([]byte, c.maxMessageSize)
	count, err := c.connection.Read(response)
	if err != nil {
		return nil, fmt.Errorf("read from connection: %w", err)
	}

	return response[:count], nil
}

func (c *TCPClient) Close() error {
	return c.connection.Close()
}
