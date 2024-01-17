package config

import (
	"context"
	"time"

	"github.com/muonsoft/validation"
	"github.com/muonsoft/validation/it"
	"github.com/spf13/afero"
)

const (
	DefaultAddress        = "localhost:3434"
	DefaultMaxMessageSize = 10_000
	DefaultMaxConnections = 100
	DefaultIdleTimeout    = time.Minute

	DefaultWALFlushingBatchSize    = 100
	DefaultWALFlushingBatchTimeout = 20 * time.Millisecond
	DefaultWALMaxSegmentSize       = 4 * 1024 * 1024
)

func DefaultServerOptions() *ServerOptions {
	return &ServerOptions{
		Engine: Engine{
			Type: "in_memory",
		},
		WAL: WAL{
			Enabled:              true,
			FlushingBatchSize:    DefaultWALFlushingBatchSize,
			FlushingBatchTimeout: DefaultWALFlushingBatchTimeout,
			MaxSegmentSize:       DefaultWALMaxSegmentSize,
			DataDirectory:        "/wal",
		},
		Network: Network{
			Address:        DefaultAddress,
			MaxConnections: DefaultMaxConnections,
			MaxMessageSize: DefaultMaxMessageSize,
			IdleTimeout:    DefaultIdleTimeout,
		},
		Logging: Logging{
			Level:  "info",
			Output: "stdout",
		},
	}
}

type ServerOptions struct {
	FS      afero.Fs
	Engine  Engine
	WAL     WAL
	Network Network
	Logging Logging
}

func (p *ServerOptions) Validate(ctx context.Context, validator *validation.Validator) error {
	return validator.Validate(ctx,
		validation.ValidProperty("engine", p.Engine),
		validation.ValidProperty("wal", p.WAL),
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

type WAL struct {
	Enabled              bool
	FlushingBatchSize    int
	FlushingBatchTimeout time.Duration
	MaxSegmentSize       int
	DataDirectory        string
}

func (w WAL) Validate(ctx context.Context, validator *validation.Validator) error {
	return validator.Validate(ctx,
		validation.NumberProperty("flushingBatchSize", w.FlushingBatchSize, it.IsBetween(1, 10_000)),
		validation.NumberProperty("flushingBatchTimeout", w.FlushingBatchTimeout, it.IsBetween(time.Millisecond, time.Hour)),
		validation.NumberProperty(
			"maxSegmentSize", w.MaxSegmentSize,
			it.IsBetween(
				100*1024,      // 100 KB
				100*1024*1024, // 100 MB
			),
		),
	)
}

type Network struct {
	Address        string
	MaxConnections int
	MaxMessageSize int
	IdleTimeout    time.Duration
	OnServerStart  func()
}

func (n Network) Validate(ctx context.Context, validator *validation.Validator) error {
	return validator.Validate(ctx,
		validation.StringProperty(
			"address", n.Address,
			it.IsNotBlank(),
		),
		validation.NumberProperty(
			"max_connections", n.MaxConnections,
			it.IsBetween(1, 10_000),
		),
	)
}

type Logging struct {
	Level  string
	Output string
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
