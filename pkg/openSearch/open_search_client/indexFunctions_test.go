package open_search_client

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndexFunction(t *testing.T) {
	ctx := context.Background()
	opensearchContainer, conf, err := StartOpensearchTestContainer(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, opensearchContainer)
	defer func() {
		opensearchContainer.Terminate(ctx)
	}()

	// Init OpenSearch
	client, err := NewOpensearchProjectClient(context.Background(), conf)
	require.NoError(t, err)

	// Init Index
	// TODO schema, err := migrations.GetIndexSchema()
	assert.NoError(t, err)
	iFunc := NewIndexFunction(client)

	// create schema

	t.Run("Test", func(t *testing.T) {
		doesNotExists, _ := iFunc.IndexExists("test")
		assert.Equal(t, false, doesNotExists)

		existingAlias, err := iFunc.AliasExists("go_vulnerability")
		assert.Nil(t, err)
		assert.Equal(t, false, existingAlias)

		// TODO add tests
	})
}
