package config

import (
	"context"
	"errors"
	"fmt"

	"github.com/dustin/go-humanize"
	"github.com/muonsoft/validation/validator"
	"github.com/spf13/viper"
)

func LoadServerOptions() (*ServerOptions, error) {
	loader := viper.New()
	loader.SetConfigName("kvdb")
	loader.SetConfigType("yaml")
	loader.AddConfigPath(".")
	if err := loader.ReadInConfig(); err != nil {
		if errors.As(err, &viper.ConfigFileNotFoundError{}) {
			return initServerOptions(loader)
		}

		return &ServerOptions{}, fmt.Errorf("read config: %w", err)
	}

	options, err := loadServerOptions(loader)
	if err != nil {
		return nil, err
	}

	if err := validator.ValidateIt(context.Background(), options); err != nil {
		return nil, err
	}

	return options, nil
}

func initServerOptions(loader *viper.Viper) (*ServerOptions, error) {
	options := DefaultServerOptions()

	loader.Set("engine.type", options.Engine.Type)
	loader.Set("wal.enabled", options.WAL.Enabled)
	loader.Set("wal.flushing_batch_size", options.WAL.FlushingBatchSize)
	loader.Set("wal.flushing_batch_timeout", options.WAL.FlushingBatchTimeout)
	loader.Set("wal.max_segment_size", humanize.Bytes(uint64(options.WAL.MaxSegmentSize)))
	loader.Set("wal.data_directory", options.WAL.DataDirectory)
	loader.Set("network.address", options.Network.Address)
	loader.Set("network.max_connections", options.Network.MaxConnections)
	loader.Set("network.max_message_size", humanize.Bytes(uint64(options.Network.MaxMessageSize)))
	loader.Set("network.idle_timeout", options.Network.IdleTimeout)
	loader.Set("logging.level", options.Logging.Level)
	loader.Set("logging.output", options.Logging.Output)
	if err := loader.SafeWriteConfig(); err != nil {
		return nil, fmt.Errorf("init config: %w", err)
	}

	return options, nil
}

func loadServerOptions(loader *viper.Viper) (*ServerOptions, error) {
	errs := make([]error, 0)

	maxMessageSize, err := humanize.ParseBytes(loader.GetString("network.max_message_size"))
	if err != nil {
		errs = append(errs, fmt.Errorf(`parse "network.max_message_size": %w`, err))
	}
	walMaxSegmentSize, err := humanize.ParseBytes(loader.GetString("wal.max_segment_size"))
	if err != nil {
		errs = append(errs, fmt.Errorf(`parse "wal.max_segment_size": %w`, err))
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return &ServerOptions{
		Engine: Engine{
			Type: loader.GetString("engine.type"),
		},
		WAL: WAL{
			Enabled:              loader.GetBool("wal.enabled"),
			FlushingBatchSize:    loader.GetInt("wal.flushing_batch_size"),
			FlushingBatchTimeout: loader.GetDuration("wal.flushing_batch_timeout"),
			MaxSegmentSize:       int(walMaxSegmentSize),
			DataDirectory:        loader.GetString("wal.data_directory"),
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
