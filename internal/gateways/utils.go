package gateways

import (
	"context"

	"github.com/wakeup-labs/issuer-node/internal/core/domain"
	"github.com/wakeup-labs/issuer-node/internal/eth"
	"github.com/wakeup-labs/issuer-node/internal/log"
	"github.com/wakeup-labs/issuer-node/internal/network"
)

func getEthClient(ctx context.Context, identity *domain.Identity, resolver network.Resolver) (*eth.Client, error) {
	resolverPrefix, err := identity.GetResolverPrefix()
	if err != nil {
		log.Error(ctx, "failed to get networkResolver prefix", "err", err)
		return nil, err
	}

	client, err := resolver.GetEthClient(resolverPrefix)
	if err != nil {
		log.Error(ctx, "failed to get client", "err", err)
		return nil, err
	}

	return client, nil
}
