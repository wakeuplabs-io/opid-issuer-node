package revocationstatus

import (
	"context"
	"testing"

	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"
	"github.com/stretchr/testify/require"
	"github.com/wakeup-labs/issuer-node/internal/common"
	"github.com/wakeup-labs/issuer-node/internal/config"
	"github.com/wakeup-labs/issuer-node/internal/network"
)

func TestRevocationStatusResolver_GetCredentialRevocationStatus(t *testing.T) {
	const did = "did:opid:optimism:sepolia:2qFbNk3Vz7Uy3ryq6zjwkC7p7RbLTfRpMsy6axjxeG"
	didW3c, err := w3c.ParseDID(did)
	require.NoError(t, err)

	type expected struct {
		err error
		*verifiable.CredentialStatus
	}

	type testConfig struct {
		name                     string
		credentialStatusSettings network.RhsSettings
		credentialStatusType     verifiable.CredentialStatusType
		nonce                    uint64
		issuerState              string
		expected                 expected
	}

	for _, tc := range []testConfig{
		{
			name: "Iden3ReverseSparseMerkleTreeProof for single issuer",
			credentialStatusSettings: network.RhsSettings{
				Mode:                 network.OffChain,
				Iden3CommAgentStatus: "https://issuernode",
				SingleIssuer:         true,
			},
			credentialStatusType: verifiable.Iden3ReverseSparseMerkleTreeProof,
			nonce:                12345,
			issuerState:          "issuer-state",
			expected: expected{
				err: nil,
				CredentialStatus: &verifiable.CredentialStatus{
					Type:            verifiable.Iden3ReverseSparseMerkleTreeProof,
					ID:              "https://rhs-staging.polygonid.me/node?state=issuer-state",
					RevocationNonce: 12345,
					StatusIssuer: &verifiable.CredentialStatus{
						Type:            verifiable.Iden3commRevocationStatusV1,
						ID:              "https://issuer-node.privado.id/v2/agent",
						RevocationNonce: 12345,
					},
				},
			},
		},
		{
			name: "Iden3ReverseSparseMerkleTreeProof for multiples issuers",
			credentialStatusSettings: network.RhsSettings{
				Mode:                 network.OffChain,
				Iden3CommAgentStatus: "https://issuernode",
				SingleIssuer:         true,
			},
			credentialStatusType: verifiable.Iden3ReverseSparseMerkleTreeProof,
			nonce:                12345,
			issuerState:          "issuer-state",
			expected: expected{
				err: nil,
				CredentialStatus: &verifiable.CredentialStatus{
					Type:            verifiable.Iden3ReverseSparseMerkleTreeProof,
					ID:              "https://rhs-staging.polygonid.me/node?state=issuer-state",
					RevocationNonce: 12345,
					StatusIssuer: &verifiable.CredentialStatus{
						Type:            verifiable.Iden3commRevocationStatusV1,
						ID:              "https://issuer-node.privado.id/v2/agent",
						RevocationNonce: 12345,
					},
				},
			},
		},
		{
			name: "Iden3OnchainSparseMerkleTreeProof2023 for single issuer",
			credentialStatusSettings: network.RhsSettings{
				Mode:                 network.OnChain,
				Iden3CommAgentStatus: "https://issuernode",
				SingleIssuer:         true,
				RhsUrl:               common.ToPointer("https://rhs"),
				ContractAddress:      common.ToPointer("0x1234567890"),
				PublishingKey:        "pbkey",
				ChainID:              common.ToPointer("80002"),
			},
			credentialStatusType: verifiable.Iden3OnchainSparseMerkleTreeProof2023,
			nonce:                12345,
			issuerState:          "issuer-state",
			expected: expected{
				err: nil,
				CredentialStatus: &verifiable.CredentialStatus{
					Type:            verifiable.Iden3OnchainSparseMerkleTreeProof2023,
					ID:              "did:opid:optimism:sepolia:2qFbNk3Vz7Uy3ryq6zjwkC7p7RbLTfRpMsy6axjxeG/credentialStatus?revocationNonce=12345&contractAddress=80001:0x0000000000000000000000000000001234567890&state=issuer-state",
					RevocationNonce: 12345,
				},
			},
		},
		{
			name: "Iden3OnchainSparseMerkleTreeProof2023 for multiples issuers",
			credentialStatusSettings: network.RhsSettings{
				Mode:                 network.OnChain,
				Iden3CommAgentStatus: "https://issuernode",
				SingleIssuer:         true,
				RhsUrl:               common.ToPointer("https://rhs"),
				ContractAddress:      common.ToPointer("0x1234567890"),
				PublishingKey:        "pbkey",
				ChainID:              common.ToPointer("80002"),
			},
			credentialStatusType: verifiable.Iden3OnchainSparseMerkleTreeProof2023,
			nonce:                12345,
			issuerState:          "issuer-state",
			expected: expected{
				err: nil,
				CredentialStatus: &verifiable.CredentialStatus{
					Type:            verifiable.Iden3OnchainSparseMerkleTreeProof2023,
					ID:              "did:opid:optimism:sepolia:2qFbNk3Vz7Uy3ryq6zjwkC7p7RbLTfRpMsy6axjxeG/credentialStatus?revocationNonce=12345&contractAddress=80001:0x0000000000000000000000000000001234567890&state=issuer-state",
					RevocationNonce: 12345,
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Configuration{
				ServerUrl:           "https://issuer-node.privado.id",
				NetworkResolverPath: "",
			}
			networkResolver, err := network.NewResolver(context.Background(), *cfg, nil, common.CreateFile(t))
			require.NoError(t, err)
			rsr := NewRevocationStatusResolver(*networkResolver)
			credentialStatus, err := rsr.GetCredentialRevocationStatus(context.Background(), *didW3c, tc.nonce, tc.issuerState, tc.credentialStatusType)
			require.Equal(t, tc.expected.CredentialStatus, credentialStatus)
			require.NoError(t, err)
		})
	}
}
