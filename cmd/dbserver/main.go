package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/strider2038/key-value-database/internal/config"
	"github.com/strider2038/key-value-database/internal/di"
)

func main() {
	options, err := config.LoadServerOptions()
	if err != nil {
		log.Fatalln("load config:", err)
	}

	if err := runServer(options); err != nil {
		log.Fatalln(err)
	}
}

func runServer(options *config.ServerOptions) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	server, err := di.NewServer(options)
	if err != nil {
		return fmt.Errorf("create server: %w", err)
	}

	if err := server.Serve(ctx); err != nil {
		return fmt.Errorf("run server: %w", err)
	}

	return nil
}
