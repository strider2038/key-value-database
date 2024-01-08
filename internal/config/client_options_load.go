package config

import (
	"fmt"

	"github.com/dustin/go-humanize"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func LoadClientOptions() (ClientOptions, error) {
	pflag.StringP("address", "a", DefaultAddress, "Database network address.")
	pflag.String("max-message-size", humanize.Bytes(DefaultMaxMessageSize), "Max message size, example: 100 Kb.")
	pflag.Duration("idle-timeout", DefaultIdleTimeout, "Network idle timeout, example: 10 s.")
	pflag.Parse()
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		return ClientOptions{}, err
	}

	maxMessageSize, err := humanize.ParseBytes(viper.GetString("max-message-size"))
	if err != nil {
		return ClientOptions{}, fmt.Errorf(`parse "max-message-size": %w`, err)
	}

	return ClientOptions{
		Address:        viper.GetString("address"),
		MaxMessageSize: int(maxMessageSize),
		IdleTimeout:    viper.GetDuration("idle-timeout"),
	}, nil
}
