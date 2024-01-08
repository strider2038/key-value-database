package config

import (
	"context"
	"errors"
	"fmt"

	"github.com/dustin/go-humanize"
	"github.com/muonsoft/validation/validator"
	"github.com/spf13/viper"
)

func LoadServerOptions() (ServerOptions, error) {
	loader := viper.New()
	loader.SetConfigName("kvdb")
	loader.SetConfigType("yaml")
	loader.AddConfigPath(".")
	if err := loader.ReadInConfig(); err != nil {
		if errors.As(err, &viper.ConfigFileNotFoundError{}) {
			return initServerOptions(loader)
		}

		return ServerOptions{}, fmt.Errorf("read config: %w", err)
	}

	options, err := loadServerOptions(loader)
	if err != nil {
		return ServerOptions{}, err
	}

	if err := validator.ValidateIt(context.Background(), options); err != nil {
		return ServerOptions{}, err
	}

	return options, nil
}

func initServerOptions(loader *viper.Viper) (ServerOptions, error) {
	options := DefaultServerOptions()

	loader.Set("engine.type", options.Engine.Type)
	loader.Set("network.address", options.Network.Address)
	loader.Set("network.max_connections", options.Network.MaxConnections)
	loader.Set("network.max_message_size", humanize.Bytes(uint64(options.Network.MaxMessageSize)))
	loader.Set("network.idle_timeout", options.Network.IdleTimeout)
	loader.Set("logging.level", options.Logging.Level)
	loader.Set("logging.output", options.Logging.Output)
	if err := loader.SafeWriteConfig(); err != nil {
		return ServerOptions{}, fmt.Errorf("init config: %w", err)
	}

	return options, nil
}

func loadServerOptions(loader *viper.Viper) (ServerOptions, error) {
	maxMessageSize, err := humanize.ParseBytes(loader.GetString("network.max_message_size"))
	if err != nil {
		return ServerOptions{}, fmt.Errorf(`parse "network.max_message_size": %w`, err)
	}

	return ServerOptions{
		Engine: Engine{
			Type: loader.GetString("engine.type"),
		},
		Network: Network{
			Address:        loader.GetString("network.address"),
			MaxConnections: loader.GetInt("network.max_connections"),
			MaxMessageSize: int(maxMessageSize),
			IdleTimeout:    loader.GetDuration("network.idle_timeout"),
		},
		Logging: Logging{
			Level:  loader.GetString("logging.level"),
			Output: loader.GetString("logging.output"),
		},
	}, nil
}
