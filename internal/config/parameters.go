package config

import (
	"context"

	"github.com/muonsoft/validation"
	"github.com/muonsoft/validation/it"
)

func Default() Parameters {
	return Parameters{
		Engine: Engine{
			Type: "in_memory",
		},
		Network: Network{
			Address:        "localhost",
			MaxConnections: 100,
		},
		Logging: Logging{
			Level:  "info",
			Output: "stdout",
		},
	}
}

type Parameters struct {
	Engine  Engine  `mapstructure:"engine"`
	Network Network `mapstructure:"network"`
	Logging Logging `mapstructure:"logging"`
}

func (p Parameters) Validate(ctx context.Context, validator *validation.Validator) error {
	return validator.Validate(ctx,
		validation.ValidProperty("engine", p.Engine),
		validation.ValidProperty("network", p.Network),
		validation.ValidProperty("logging", p.Logging),
	)
}

type Engine struct {
	Type string `yaml:"type"`
}

func (e Engine) Validate(ctx context.Context, validator *validation.Validator) error {
	return validator.Validate(ctx,
		validation.StringProperty(
			"type", e.Type,
			it.IsOneOf("in_memory").WithMessage("Must be one of: {{ choices }}."),
		),
	)
}

type Network struct {
	Address        string `mapstructure:"address"`
	MaxConnections int    `mapstructure:"max_connections"`
}

func (n Network) Validate(ctx context.Context, validator *validation.Validator) error {
	return validator.Validate(ctx,
		validation.StringProperty("address", n.Address, it.IsNotBlank()),
		validation.NumberProperty("max_connections", n.MaxConnections, it.IsBetween(1, 10_000)),
	)
}

type Logging struct {
	Level  string `mapstructure:"level"`
	Output string `mapstructure:"output"`
}

func (l Logging) Validate(ctx context.Context, validator *validation.Validator) error {
	return validator.Validate(ctx,
		validation.StringProperty(
			"level", l.Level,
			it.IsOneOf("debug", "info", "warn", "warning", "error").
				WithMessage("Must be one of: {{ choices }}."),
		),
	)
}
