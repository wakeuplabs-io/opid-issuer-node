package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/wakeup-labs/issuer-node/internal/core/domain"
	"github.com/wakeup-labs/issuer-node/internal/db"
)

// ConnectionRepository defines the available methods for connections repository
type ConnectionRepository interface {
	Save(ctx context.Context, conn db.Querier, connection *domain.Connection) (uuid.UUID, error)
	Delete(ctx context.Context, conn db.Querier, id uuid.UUID, issuerDID w3c.DID) error
	DeleteCredentials(ctx context.Context, conn db.Querier, id uuid.UUID, issuerID w3c.DID) error
	GetByIDAndIssuerID(ctx context.Context, conn db.Querier, id uuid.UUID, issuerDID w3c.DID) (*domain.Connection, error)
	GetByUserID(ctx context.Context, conn db.Querier, issuerDID w3c.DID, userDID w3c.DID) (*domain.Connection, error)
	GetAllWithCredentialsByIssuerID(ctx context.Context, conn db.Querier, issuerDID w3c.DID, filter *NewGetAllConnectionsRequest) ([]domain.Connection, uint, error)
	GetByUserSessionID(ctx context.Context, conn db.Querier, sessionID uuid.UUID) (*domain.Connection, error)
	SaveUserAuthentication(ctx context.Context, conn db.Querier, connID uuid.UUID, sessID uuid.UUID, mTime time.Time) error
}
