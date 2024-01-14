package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/strider2038/key-value-database/internal/database/engine"
	"github.com/strider2038/key-value-database/internal/database/network"
)

type Network interface {
	Serve(ctx context.Context, handler network.Handler) error
}

type Server struct {
	controller *engine.Controller
	network    Network
	logger     *slog.Logger
}

func NewServer(
	controller *engine.Controller,
	network Network,
	logger *slog.Logger,
) *Server {
	return &Server{
		controller: controller,
		network:    network,
		logger:     logger,
	}
}

func (s *Server) Serve(ctx context.Context) error {
	if err := s.network.Serve(ctx, network.HandlerFunc(s.handleRequest)); err != nil {
		return fmt.Errorf("serve: %w", err)
	}

	return nil
}

func (s *Server) handleRequest(ctx context.Context, request []byte) []byte {
	response, err := s.controller.Execute(ctx, string(request))
	if err != nil {
		var badRequest *engine.BadRequestError
		if errors.As(err, &badRequest) {
			return []byte(fmt.Sprintf("Bad request: %s", badRequest.Unwrap()))
		}

		s.logger.Error("Internal server error", "error", err)

		return []byte("Internal server error")
	}

	return []byte(response)
}
