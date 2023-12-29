package database_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/strider2038/key-value-database/internal/database/computation/basic/analyzing"
	"github.com/strider2038/key-value-database/internal/di"
)

type DatabaseTestStep struct {
	Request      string
	WantResponse string
	WantError    error
}

func TestDatabase_Execute(t *testing.T) {
	tests := []struct {
		name  string
		steps []DatabaseTestStep
	}{
		{
			name: "set - get",
			steps: []DatabaseTestStep{
				{
					Request:      "SET key value",
					WantResponse: "OK",
				},
				{
					Request:      "GET key",
					WantResponse: "value",
				},
			},
		},
		{
			name: "set - get - del - get",
			steps: []DatabaseTestStep{
				{
					Request:      "SET key value",
					WantResponse: "OK",
				},
				{
					Request:      "GET key",
					WantResponse: "value",
				},
				{
					Request:      "DEL key",
					WantResponse: "OK",
				},
				{
					Request:      "GET key",
					WantResponse: "$_",
				},
			},
		},
		{
			name: "get not found",
			steps: []DatabaseTestStep{
				{
					Request:      "GET key",
					WantResponse: "$_",
				},
			},
		},
		{
			name: "invalid command",
			steps: []DatabaseTestStep{
				{
					Request:   "GET",
					WantError: analyzing.ErrNotEnoughArguments,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := di.NewDatabase(di.Options{})

			for i, step := range test.steps {
				response, err := db.Execute(context.Background(), step.Request)

				if step.WantError == nil {
					require.NoError(t, err, "step %d", i)
					assert.Equal(t, step.WantResponse, response, "step %d", i)
				} else {
					assert.ErrorIs(t, err, step.WantError, "step %d", i)
				}
			}
		})
	}
}
