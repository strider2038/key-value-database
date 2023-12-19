package database_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/strider2038/key-value-database/internal/database"
	"github.com/strider2038/key-value-database/internal/database/computing"
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
					WantResponse: "VALUE value",
				},
				{
					Request:      "DEL key",
					WantResponse: "OK",
				},
				{
					Request:      "GET key",
					WantResponse: "NOT FOUND",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := database.NewDatabase(computing.NewComputer())

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
