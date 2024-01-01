package main

import (
	"context"
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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	server, err := di.NewServer(options)
	if err != nil {
		log.Fatalln("create server: ", err)
	}

	if err := server.Serve(ctx); err != nil {
		log.Fatalln("run server: ", err)
	}
}
