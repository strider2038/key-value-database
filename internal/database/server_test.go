package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/strider2038/key-value-database/internal/config"
	"github.com/strider2038/key-value-database/internal/database/network"
	"github.com/strider2038/key-value-database/internal/di"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

type ServerTestStep struct {
	Request      string
	WantResponse string
}

func TestServer_Serve(t *testing.T) {
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

	const address = ":11000"

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, stop := context.WithCancel(context.Background())
			defer stop()
			waitServer := make(chan struct{})
			waitFinish := make(chan struct{})

			server, err := di.NewServer(config.ServerOptions{
				Network: config.Network{
					Address:        address,
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
			client, err := network.NewTCPClient(address, 10_000, time.Second)
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

func waitSecond(tb testing.TB, wait chan struct{}) {
	tb.Helper()
	select {
	case <-wait:
	case <-time.After(time.Second):
		assert.FailNow(tb, "waiting on channel")
	}
}
