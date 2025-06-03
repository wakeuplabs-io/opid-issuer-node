package services

import (
	"context"

	core "github.com/iden3/go-iden3-core/v2"
	"github.com/wakeup-labs/issuer-node/internal/log"
)

const (
	// DIDMethodOptimismID Optimism DID method
	DIDMethodOptimismID core.DIDMethod = "opid"
	// DIDMethodOptimismByte Optimism DID method byte
	DIDMethodOptimismByte byte = 0b00000011
	// Optimism blockchain identifier
	Optimism core.Blockchain = "optimism"
	// OptimismChainID Optimism mainnet chain id
	OptimismChainID = 10
	// OptimismSepoliaChainID Optimism sepolia chain id
	OptimismSepoliaChainID = 11155420
	// OptimismNetworkFlag Optimism network flag
	OptimismNetworkFlag = 0b1000_0000 | 0b0000_0001
	// OptimismNetworkFlagSepolia Optimism sepolia network flag
	OptimismNetworkFlagSepolia = 0b1000_0000 | 0b0000_0010
)

// RegisterOptimismIdMethod registers Optimism DID method
func RegisterOptimismIdMethod(ctx context.Context) error {
	// register did method
	if err := core.RegisterDIDMethod(DIDMethodOptimismID, DIDMethodOptimismByte); err != nil {
		log.Error(ctx, "cannot register opid method", "err", err)
		return err
	}

	// register sepolia network
	sepoliaParams := core.DIDMethodNetworkParams{
		Method:      DIDMethodOptimismID,
		Blockchain:  Optimism,
		Network:     core.Sepolia,
		NetworkFlag: OptimismNetworkFlagSepolia, // chain | network (0b0000_0010 used for testnet usually)
	}
	if err := core.RegisterDIDMethodNetwork(sepoliaParams, core.WithChainID(OptimismSepoliaChainID)); err != nil {
		log.Error(ctx, "cannot register opid network", "err", err)
		return err
	}

	// register mainnet network
	mainnetParams := core.DIDMethodNetworkParams{
		Method:      DIDMethodOptimismID,
		Blockchain:  Optimism,
		Network:     core.Main,
		NetworkFlag: OptimismNetworkFlag, // chain | network (0b0000_0001 used for main usually)
	}
	if err := core.RegisterDIDMethodNetwork(mainnetParams, core.WithChainID(OptimismChainID)); err != nil {
		log.Error(ctx, "cannot register opid network", "err", err)
		return err
	}

	return nil
}
