package config

import (
	"context"
	"errors"
	"fmt"

	"github.com/muonsoft/validation/validator"
	"github.com/spf13/viper"
)

func Load() (Parameters, error) {
	loader := viper.New()
	loader.SetConfigName("kvdb")
	loader.SetConfigType("yaml")
	loader.AddConfigPath(".")
	if err := loader.ReadInConfig(); err != nil {
		if errors.As(err, &viper.ConfigFileNotFoundError{}) {
			return initConfig(loader)
		}

		return Parameters{}, fmt.Errorf("read config: %w", err)
	}

	params := Parameters{}
	if err := loader.Unmarshal(&params); err != nil {
		return Parameters{}, fmt.Errorf("parse config: %w", err)
	}

	if err := validator.ValidateIt(context.Background(), params); err != nil {
		return Parameters{}, err
	}

	return params, nil
}

func initConfig(loader *viper.Viper) (Parameters, error) {
	params := Default()

	loader.Set("engine.type", params.Engine.Type)
	loader.Set("network.address", params.Network.Address)
	loader.Set("network.max_connections", params.Network.MaxConnections)
	loader.Set("logging.level", params.Logging.Level)
	loader.Set("logging.output", params.Logging.Output)
	if err := loader.SafeWriteConfig(); err != nil {
		return Parameters{}, fmt.Errorf("init config: %w", err)
	}

	return params, nil
}
