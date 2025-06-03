package repositories

import (
	"context"
	"testing"

	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/stretchr/testify/assert"
)

func TestMtSave(t *testing.T) {
	// given
	idStr := "did:opid:optimism:sepolia:wyFiV4w71QgWPn6bYLsZoysFay66gKtVa9kfu6yMZ"
	repo := NewIdentityMerkleTreeRepository()

	// when and then
	t.Run("should save the mt", func(t *testing.T) {
		mt, err := repo.Save(context.Background(), storage.Pgx, idStr, 0)
		assert.NoError(t, err)
		assert.NotNil(t, mt)
	})
}

func TestMtGetByIdentifierAndTypes(t *testing.T) {
	// given
	idStr := "did:opid:optimism:sepolia:2qF6oxuF6HhD45o5E1yF1gq1vdTAGtTfGqQ7bUaKeC"
	repo := NewIdentityMerkleTreeRepository()

	// when and then
	t.Run("should get the mt", func(t *testing.T) {
		mt, err := repo.Save(context.Background(), storage.Pgx, idStr, 0)
		assert.NoError(t, err)
		assert.NotNil(t, mt)
		did, err := w3c.ParseDID(idStr)
		assert.NoError(t, err)

		mts, err := repo.GetByIdentifierAndTypes(context.Background(), storage.Pgx, did, []uint16{0})
		assert.NoError(t, err)
		assert.NotNil(t, mts)
		assert.Equal(t, 1, len(mts))
	})

	// when and then
	t.Run("should not get the mt", func(t *testing.T) {
		did, err := w3c.ParseDID("did:opid:optimism:sepolia:2qHtzzxS7uazdumnyZEdf74CNo3MptdW6ytxxwbPMW")
		assert.NoError(t, err)
		mts, err := repo.GetByIdentifierAndTypes(context.Background(), storage.Pgx, did, []uint16{0})
		assert.NoError(t, err)
		assert.NotNil(t, mts)
		assert.Equal(t, 0, len(mts))
	})
}

func TestMtGetById(t *testing.T) {
	// given
	idStr := "did:opid:optimism:sepolia:2qPcZy8C1Nnm9xmkQsZZjRQ11V2YJ6VYULpg4VcxXm"
	repo := NewIdentityMerkleTreeRepository()

	// when and then
	t.Run("should get the mt", func(t *testing.T) {
		mt, err := repo.Save(context.Background(), storage.Pgx, idStr, 0)
		assert.NoError(t, err)
		assert.NotNil(t, mt)

		mts, err := repo.GetByID(context.Background(), storage.Pgx, mt.ID)
		assert.NoError(t, err)
		assert.NotNil(t, mts)
		assert.Equal(t, mt.ID, mts.ID)
	})

	// when and then
	t.Run("should not get the mt", func(t *testing.T) {
		mts, err := repo.GetByID(context.Background(), storage.Pgx, uint64(11))
		assert.Error(t, err)
		assert.Nil(t, mts)
	})
}
