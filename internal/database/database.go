package database

import (
	"context"

	"github.com/strider2038/key-value-database/internal/database/querylang"
)

type Computer interface {
	ParseRequest(ctx context.Context, request string) (*querylang.Command, error)
}

type Database struct {
	computer Computer
}

func NewDatabase(computer Computer) *Database {
	return &Database{computer: computer}
}

func (db *Database) Execute(ctx context.Context, request string) (string, error) {
	return "", nil
}
