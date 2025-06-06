package repositories

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/wakeup-labs/issuer-node/internal/core/domain"
)

func TestGetIdentities(t *testing.T) {
	fixture := NewFixture(storage)
	idStr1 := "did:opid:optimism:sepolia:2qGqLpDT2VyqFq1NmfRkB9gwLxBhMRuazv2ZgHfjUw"
	idStr2 := "did:opid:optimism:sepolia:2qNR5sUiiSt5v6bnKQZyjCu2n9uNbKD34cZkSkgwUq"

	identity1 := &domain.Identity{
		Identifier: idStr1,
	}

	identity2 := &domain.Identity{
		Identifier: idStr2,
	}
	fixture.CreateIdentity(t, identity1)
	fixture.CreateIdentity(t, identity2)

	identityRepo := NewIdentity()
	t.Run("should get identities", func(t *testing.T) {
		identities, err := identityRepo.Get(context.Background(), storage.Pgx)
		assert.NoError(t, err)
		assert.True(t, len(identities) >= 2)
	})
}
