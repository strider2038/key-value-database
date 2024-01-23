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

type NetworkService struct {
	controller *engine.Controller
	network    Network
	logger     *slog.Logger
}

func NewNetworkService(
	controller *engine.Controller,
	network Network,
	logger *slog.Logger,
) *NetworkService {
	return &NetworkService{
		controller: controller,
		network:    network,
		logger:     logger,
	}
}

func (s *NetworkService) Serve(ctx context.Context) error {
	if err := s.network.Serve(ctx, network.HandlerFunc(s.handleRequest)); err != nil {
		return fmt.Errorf("serve: %w", err)
	}

	return nil
}

func (s *NetworkService) handleRequest(ctx context.Context, request []byte) []byte {
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
