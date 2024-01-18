package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/strider2038/key-value-database/internal/config"
	"github.com/strider2038/key-value-database/internal/database"
	"github.com/strider2038/key-value-database/internal/database/network"
	"github.com/strider2038/key-value-database/internal/di"
	"go.uber.org/goleak"
)

const ServerAddress = ":11000"

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

type ServerTestStep struct {
	Request      string
	WantResponse string
}

func TestServer_Serve_NonPersistentMode(t *testing.T) {
	tests := []struct {
		name  string
		steps []ServerTestStep
	}{
		{
			name: "set - get",
			steps: []ServerTestStep{
				{
					Request:      "SET key value",
					WantResponse: "OK",
				},
				{
					Request:      "GET key",
					WantResponse: "value",
				},
			},
		},
		{
			name: "set - get - del - get",
			steps: []ServerTestStep{
				{
					Request:      "SET key value",
					WantResponse: "OK",
				},
				{
					Request:      "GET key",
					WantResponse: "value",
				},
				{
					Request:      "DEL key",
					WantResponse: "OK",
				},
				{
					Request:      "GET key",
					WantResponse: "$_",
				},
			},
		},
		{
			name: "get not found",
			steps: []ServerTestStep{
				{
					Request:      "GET key",
					WantResponse: "$_",
				},
			},
		},
		{
			name: "invalid command",
			steps: []ServerTestStep{
				{
					Request:      "test",
					WantResponse: "Bad request: parse command: analyze command: unknown command",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, stop := context.WithCancel(context.Background())
			defer stop()
			waitServer := make(chan struct{})
			waitFinish := make(chan struct{})

			server, err := di.NewServer(&config.ServerOptions{
				Network: config.Network{
					Address:        ServerAddress,
					MaxConnections: 1,
					MaxMessageSize: 10_000,
					IdleTimeout:    time.Second,
					OnServerStart:  func() { close(waitServer) },
				},
			})
			require.NoError(t, err)
			go func() {
				assert.NoError(t, server.Serve(ctx))
				close(waitFinish)
			}()

			waitSecond(t, waitServer)
			client, err := network.NewTCPClient(ServerAddress, 10_000, time.Second)
			require.NoError(t, err)
			defer client.Close()

			for i, step := range test.steps {
				response, err := client.Send([]byte(step.Request))
				require.NoError(t, err, "step %d", i)
				assert.Equal(t, step.WantResponse, string(response), "step %d", i)
			}

			client.Close()
			stop()
			waitSecond(t, waitFinish)
		})
	}
}

func TestServer_Serve_WithWAL(t *testing.T) {
	waitServer := make(chan struct{})
	waitFinish := make(chan struct{})

	fs := afero.NewMemMapFs()

	// Запускаем сервер первый раз
	server := createServerWithWAL(t, fs, waitServer)
	ctx, stop := context.WithCancel(context.Background())
	go func() {
		assert.NoError(t, server.Serve(ctx))
		waitFinish <- struct{}{}
	}()

	// Отправляем команды записи на сервер БД
	waitSecond(t, waitServer)
	sendCommandsToServer(t, []ServerTestStep{
		{Request: "SET key value", WantResponse: "OK"},
		{Request: "SET foo 1", WantResponse: "OK"},
		{Request: "SET bar 2", WantResponse: "OK"},
		{Request: "DEL key", WantResponse: "OK"},
		{Request: "SET baz 3", WantResponse: "OK"},
	})

	// Останавливаем сервер
	stop()
	waitSecond(t, waitFinish)

	// Перезапускаем сервер, который должен восстановить свое состояние
	server = createServerWithWAL(t, fs, waitServer)
	ctx, stop = context.WithCancel(context.Background())
	go func() {
		assert.NoError(t, server.Serve(ctx))
		waitFinish <- struct{}{}
	}()

	// Отправляем команды на чтение для верификации восстановленного состояния
	waitSecond(t, waitServer)
	sendCommandsToServer(t, []ServerTestStep{
		{Request: "GET key", WantResponse: "$_"},
		{Request: "GET foo", WantResponse: "1"},
		{Request: "GET bar", WantResponse: "2"},
		{Request: "GET baz", WantResponse: "3"},
	})

	// Останавливаем сервер
	stop()
	waitSecond(t, waitFinish)
}

func createServerWithWAL(tb testing.TB, fs afero.Fs, wait chan<- struct{}) *database.Server {
	tb.Helper()

	server, err := di.NewServer(&config.ServerOptions{
		FS: fs,
		Network: config.Network{
			Address:        ServerAddress,
			MaxConnections: 1,
			MaxMessageSize: 10_000,
			IdleTimeout:    time.Second,
			OnServerStart:  func() { wait <- struct{}{} },
		},
		WAL: config.WAL{
			Enabled:              true,
			FlushingBatchSize:    10,
			FlushingBatchTimeout: time.Millisecond,
			MaxSegmentSize:       config.DefaultWALMaxSegmentSize,
			DataDirectory:        "/wal",
		},
	})
	require.NoError(tb, err)

	return server
}

func sendCommandsToServer(tb testing.TB, writeSteps []ServerTestStep) {
	tb.Helper()

	client, err := network.NewTCPClient(ServerAddress, 10_000, time.Second)
	require.NoError(tb, err)
	defer client.Close()

	for i, step := range writeSteps {
		response, err := client.Send([]byte(step.Request))
		require.NoError(tb, err, "step %d", i)
		assert.Equal(tb, step.WantResponse, string(response), "step %d", i)
	}
}

func waitSecond(tb testing.TB, wait <-chan struct{}) {
	tb.Helper()
	select {
	case <-wait:
	case <-time.After(time.Second):
		assert.FailNow(tb, "waiting on channel")
	}
}
