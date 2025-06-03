package repositories

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/wakeup-labs/issuer-node/internal/core/ports"
	"github.com/wakeup-labs/issuer-node/internal/db"
)

// Fixture - Handle testing fixture configuration
type Fixture struct {
	storage                 *db.Storage
	identityRepository      ports.IndentityRepository
	claimRepository         ports.ClaimRepository
	connectionsRepository   ports.ConnectionRepository
	schemaRepository        ports.SchemaRepository
	identityStateRepository ports.IdentityStateRepository
}

// NewFixture - constructor
func NewFixture(storage *db.Storage) *Fixture {
	return &Fixture{
		storage:                 storage,
		identityRepository:      NewIdentity(),
		claimRepository:         NewClaim(),
		connectionsRepository:   NewConnection(),
		schemaRepository:        NewSchema(*storage),
		identityStateRepository: NewIdentityState(),
	}
}

// ExecQueryParams - handle the query and the argumens for that query.
type ExecQueryParams struct {
	Query     string
	Arguments []interface{}
}

// ExecQuery - Execute a query for testing purpose.
func (f *Fixture) ExecQuery(t *testing.T, params ExecQueryParams) {
	t.Helper()
	_, err := f.storage.Pgx.Exec(context.Background(), params.Query, params.Arguments...)
	assert.NoError(t, err)
}
