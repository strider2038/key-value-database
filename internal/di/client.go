package di

import (
	"fmt"

	"github.com/strider2038/key-value-database/internal/config"
	"github.com/strider2038/key-value-database/internal/database/network"
)

func NewClient(options config.ClientOptions) (*network.TCPClient, error) {
	client, err := network.NewTCPClient(options.Address, options.MaxMessageSize, options.IdleTimeout)
	if err != nil {
		return nil, fmt.Errorf("create TCP client: %w", err)
	}

	return client, nil
}
