package config

import "time"

type ClientOptions struct {
	Address        string
	MaxMessageSize int
	IdleTimeout    time.Duration
}
