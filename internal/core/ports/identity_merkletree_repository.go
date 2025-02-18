package ports

import (
	"context"

	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/wakeup-labs/issuer-node/internal/core/domain"
	"github.com/wakeup-labs/issuer-node/internal/db"
)

// IdentityMerkleTreeRepository is the interface that defines the available methods
type IdentityMerkleTreeRepository interface {
	Save(ctx context.Context, conn db.Querier, identifier string, mtType uint16) (*domain.IdentityMerkleTree, error)
	UpdateByID(ctx context.Context, conn db.Querier, imt *domain.IdentityMerkleTree) error
	GetByID(ctx context.Context, conn db.Querier, mtID uint64) (*domain.IdentityMerkleTree, error)
	GetByIdentifierAndTypes(ctx context.Context, conn db.Querier, identifier *w3c.DID, mtTypes []uint16) ([]domain.IdentityMerkleTree, error)
}
