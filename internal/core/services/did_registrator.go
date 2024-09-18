package services

import (
	"context"

	core "github.com/iden3/go-iden3-core/v2"

	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/log"
)

const (
	// DIDMethodOptimismID Optimism DID method
	DIDMethodOptimismID core.DIDMethod = "opid"
	// DIDMethodOptimismByte Optimism DID method byte
	DIDMethodOptimismByte byte = 0b00000011
	// Optimism blockchain identifier
	Optimism core.Blockchain = "optimism"
	// OptimismChainId Optimism mainnet chain id
	OptimismChainID = 10
	// OptimismSepoliaChainID Optimism sepolia chain id
	OptimismSepoliaChainID = 11155420
)

// RegisterCustomDIDMethods registers custom DID methods
func RegisterCustomDIDMethods(ctx context.Context, customsDis []config.CustomDIDMethods) error {
	for _, cdid := range customsDis {
		params := core.DIDMethodNetworkParams{
			Method:      DIDMethodOptimismID,
			Blockchain:  core.Blockchain(cdid.Blockchain),
			Network:     core.NetworkID(cdid.Network),
			NetworkFlag: cdid.NetworkFlag,
		}
		if err := core.RegisterDIDMethodNetwork(params, core.WithChainID(cdid.ChainID)); err != nil {
			log.Error(ctx, "cannot register custom DID method", "err", err, "customDID", cdid)
			return err
		}
	}
	return nil
}

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
		NetworkFlag: 0b1000_0000 | 0b0000_0010, // chain | network (0b0000_0010 used for testnet usually)
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
		NetworkFlag: 0b1000_0000 | 0b0000_0001, // chain | network (0b0000_0001 used for main usually)
	}
	if err := core.RegisterDIDMethodNetwork(mainnetParams, core.WithChainID(OptimismChainID)); err != nil {
		log.Error(ctx, "cannot register opid network", "err", err)
		return err
	}

	return nil
}
