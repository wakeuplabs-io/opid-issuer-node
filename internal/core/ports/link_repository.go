package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/iden3comm/v2/protocol"

	"github.com/wakeup-labs/issuer-node/internal/core/domain"
	"github.com/wakeup-labs/issuer-node/internal/db"
)

// LinkRepository the interface that defines the available methods
type LinkRepository interface {
	Save(ctx context.Context, conn db.Querier, link *domain.Link) (*uuid.UUID, error)
	GetByID(ctx context.Context, issuerID w3c.DID, id uuid.UUID) (*domain.Link, error)
	GetAll(ctx context.Context, issuerDID w3c.DID, status LinkStatus, query *string) ([]*domain.Link, error)
	Delete(ctx context.Context, id uuid.UUID, issuerDID w3c.DID) error
	AddAuthorizationRequest(ctx context.Context, linkID uuid.UUID, issuerDID w3c.DID, authorizationRequest *protocol.AuthorizationRequestMessage) error
}
