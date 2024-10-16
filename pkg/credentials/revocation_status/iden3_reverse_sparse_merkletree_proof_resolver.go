package revocation_status

import (
	"fmt"

	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"

	"github.com/wakeup-labs/issuer-node/internal/config"
)

type iden3ReverseSparseMerkleTreeProofResolver struct{}

func (r *iden3ReverseSparseMerkleTreeProofResolver) resolve(credentialStatusSettings config.CredentialStatus, issuerDID w3c.DID, nonce uint64, issuerState string) *verifiable.CredentialStatus {
	return &verifiable.CredentialStatus{
		ID:              buildRHSRevocationURL(credentialStatusSettings.RHS.GetURL(), issuerState),
		Type:            verifiable.Iden3ReverseSparseMerkleTreeProof,
		RevocationNonce: nonce,
		StatusIssuer: &verifiable.CredentialStatus{
			ID:              fmt.Sprintf("%s/v1/agent", credentialStatusSettings.Iden3CommAgentStatus.GetURL()),
			Type:            verifiable.Iden3commRevocationStatusV1,
			RevocationNonce: nonce,
		},
	}
}
