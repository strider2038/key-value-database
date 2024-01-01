package network_test

import (
	"context"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/strider2038/key-value-database/internal/database/network"
)

const messageSize = 1024

func TestTCPServer_Serve(t *testing.T) {
	const address = ":10001"
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	wait := make(chan struct{})
	onStartup := func() { close(wait) }
	server := network.NewTCPServer(address, 1, messageSize, time.Second, onStartup, logger)
	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	{
		go func() {
			err := server.Serve(ctx, network.HandlerFunc(func(ctx context.Context, bytes []byte) []byte {
				return append([]byte("echo to "), bytes...)
			}))
			require.NoError(t, err, "serve")
		}()
	}

	{
		<-wait
		connection, err := net.Dial("tcp", address)
		require.NoError(t, err, "connect to TCP server")
		defer connection.Close()

		_, err = connection.Write([]byte("request"))
		require.NoError(t, err, "write payload to TCP server")

		response := make([]byte, messageSize)
		count, err := connection.Read(response)
		require.NoError(t, err, "read from connection")
		assert.Equal(t, "echo to request", string(response[:count]))
	}
}
