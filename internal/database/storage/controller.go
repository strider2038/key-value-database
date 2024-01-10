package storage

import (
	"errors"
	"fmt"

	"github.com/strider2038/key-value-database/internal/database/querylang"
)

type Controller struct {
	storage Storage
}

func NewController(storage Storage) *Controller {
	return &Controller{storage: storage}
}

func (c *Controller) Execute(command *querylang.Command) (string, error) {
	switch command.ID() {
	case querylang.CommandGet:
		return c.handleGet(command.Arguments())
	case querylang.CommandSet:
		return c.handleSet(command.Arguments())
	case querylang.CommandDel:
		return c.handleDel(command.Arguments())
	default:
		return "", fmt.Errorf("unsupported command: %s", command.ID().String())
	}
}

func (c *Controller) handleGet(arguments []string) (string, error) {
	value, err := c.storage.Get(arguments[0])
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return querylang.Nil, nil
		}

		return "", err
	}

	return value, nil
}

func (c *Controller) handleSet(arguments []string) (string, error) {
	if err := c.storage.Set(arguments[0], arguments[1]); err != nil {
		return "", err
	}

	return "OK", nil
}

func (c *Controller) handleDel(arguments []string) (string, error) {
	if err := c.storage.Del(arguments[0]); err != nil {
		return "", err
	}

	return "OK", nil
}
