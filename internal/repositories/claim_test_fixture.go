package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wakeup-labs/issuer-node/internal/common"
	"github.com/wakeup-labs/issuer-node/internal/core/domain"
	"github.com/wakeup-labs/issuer-node/internal/core/ports"
)

const (
	defualtAuthClaims = 1
)

// CreateClaim fixture
func (f *Fixture) CreateClaim(t *testing.T, claim *domain.Claim) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	id, err := f.claimRepository.Save(ctx, f.storage.Pgx, claim)
	require.NoError(t, err)
	return id
}

// CreateSchema creates an entry in schema table
func (f *Fixture) CreateSchema(t *testing.T, ctx context.Context, s *domain.Schema) {
	t.Helper()
	require.NoError(t, f.schemaRepository.Save(ctx, s))
}

// GetDefaultAuthClaimOfIssuer returns the default auth claim of an issuer just created
func (f *Fixture) GetDefaultAuthClaimOfIssuer(t *testing.T, issuerID string) *domain.Claim {
	t.Helper()
	ctx := context.Background()
	did, err := w3c.ParseDID(issuerID)
	assert.NoError(t, err)
	claims, _, err := f.claimRepository.GetAllByIssuerID(ctx, f.storage.Pgx, *did, &ports.ClaimsFilter{})
	assert.NoError(t, err)
	require.Equal(t, len(claims), defualtAuthClaims)

	return claims[0]
}

// NewClaim fixture
// nolint
func (f *Fixture) NewClaim(t *testing.T, identity string) *domain.Claim {
	t.Helper()

	claimID, err := uuid.NewUUID()
	assert.NoError(t, err)

	nonce := int64(123)
	revNonce := domain.RevNonceUint64(nonce)
	claim := &domain.Claim{
		ID:              claimID,
		Identifier:      &identity,
		Issuer:          identity,
		SchemaHash:      "ca938857241db9451ea329256b9c06e5",
		SchemaURL:       "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/auth.json-ld",
		SchemaType:      "AuthBJJCredential",
		OtherIdentifier: "did:opid:optimism:sepolia:476e4HH6dJ87f1G1EPxoMUzjGeYLHHC1ADzdjdnsge",
		Expiration:      0,
		Version:         0,
		RevNonce:        revNonce,
		CoreClaim:       domain.CoreClaim{},
		Status:          nil,
	}

	vc := verifiable.W3CCredential{
		ID:           fmt.Sprintf("http://localhost/api/v2/credentials/%s", claimID),
		Context:      []string{"https://www.w3.org/2018/credentials/v1", "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/iden3credential-v2.json-ld", "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld"},
		Type:         []string{"VerifiableCredential", "KYCAgeCredential"},
		IssuanceDate: common.ToPointer(time.Now().UTC()),
		CredentialSubject: map[string]interface{}{
			"id":           "did:opid:optimism:sepolia:476e4HH6dJ87f1G1EPxoMUzjGeYLHHC1ADzdjdnsge",
			"birthday":     19960424,
			"documentType": 2,
			"type":         "KYCAgeCredential",
		},
		CredentialStatus: verifiable.CredentialStatus{
			ID:              fmt.Sprintf("http://localhost/v2/%s/credentials/revocation/status/%d", identity, revNonce),
			Type:            "SparseMerkleTreeProof",
			RevocationNonce: uint64(revNonce),
		},
		Issuer: identity,
		CredentialSchema: verifiable.CredentialSchema{
			ID:   "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
			Type: "JsonSchemaValidator2018",
		},
		RefreshService: &verifiable.RefreshService{
			ID:   "https://refresh-service.xyz",
			Type: verifiable.Iden3RefreshService2023,
		},
	}

	err = claim.CredentialStatus.Set(vc.CredentialStatus)
	assert.NoError(t, err)

	err = claim.Data.Set(vc)
	assert.NoError(t, err)

	return claim
}
