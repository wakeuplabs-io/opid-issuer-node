package eth

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/misc/eip1559"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/iden3/contracts-abi/state/go/abi"

	"github.com/wakeup-labs/issuer-node/internal/kms"
	"github.com/wakeup-labs/issuer-node/internal/log"
)

const (
	// Eq is for "equal" result of comparison
	Eq = 0
	// Gt is for "greater" than result of comparison
	Gt = 1
	// Lt is for "less than" result of comparison
	Lt = -1

	gasPriceIncrement               = 10
	transactionUnderpricedIncrement = 30
	feeIncrement                    = 1.25
)

var (
	// ErrPrivateKeyNil when private key is nil
	ErrPrivateKeyNil = errors.New("authorized calls can't be made with empty private key")
	// ErrReceiptStatusFailed when receiving a failed transaction
	ErrReceiptStatusFailed = errors.New("receipt status is failed")
	// ErrReceiptNotReceived when unable to retrieve a transaction
	ErrReceiptNotReceived = errors.New("receipt not available")
	// ErrTransactionNotFound transaction doesn't exist on blockchain
	ErrTransactionNotFound = errors.New("transaction not found")
	// CompressedPublicKeyLength is the length of a compressed public key
	CompressedPublicKeyLength = 33
	// AwsKmsPublicKeyLength is the length of a public key from AWS KMS
	AwsKmsPublicKeyLength = 88
)

// Client is an ethereum client to call Smart Contract methods.
type Client struct {
	client *ethclient.Client
	Config *ClientConfig
	kms    *kms.KMS
}

// ClientConfig eth client config
type ClientConfig struct {
	ReceiptTimeout         time.Duration `json:"receipt_timeout"`
	ConfirmationTimeout    time.Duration `json:"confirmation_timeout"`
	ConfirmationBlockCount int64         `json:"confirmation_block_count"`
	DefaultGasLimit        int           `json:"default_gas_limit"`
	MinGasPrice            *big.Int      `json:"min_gas_price"`
	MaxGasPrice            *big.Int      `json:"max_gas_price"`
	GasLess                bool          `json:"gas_less"`
	RPCResponseTimeout     time.Duration `json:"rpc_response_time_out"`
	WaitReceiptCycleTime   time.Duration `json:"wait_receipt*eth.Client_cycle_time_out"`
	WaitBlockCycleTime     time.Duration `json:"wait_block_cycle_time_out"`
}

// NewClient creates a Client instance.
func NewClient(client *ethclient.Client, c *ClientConfig, kms *kms.KMS) *Client {
	return &Client{
		client: client,
		Config: c,
		kms:    kms,
	}
}

// GetEthereumClient returns the underlying ethereum client
func (c *Client) GetEthereumClient() *ethclient.Client {
	return c.client
}

// GetConfirmationBlockCount returns the number of blocks to wait for confirmation
func (c *Client) GetConfirmationBlockCount() int64 {
	return c.Config.ConfirmationBlockCount
}

// GetConfirmationConfirmationTimeout returns the timeout for confirmation
func (c *Client) GetConfirmationConfirmationTimeout() time.Duration {
	return c.Config.ConfirmationTimeout
}

// BalanceAt retrieves information about the default account
func (c *Client) BalanceAt(ctx context.Context, addr common.Address) (*big.Int, error) {
	_ctx, cancel := context.WithTimeout(ctx, c.Config.RPCResponseTimeout)
	defer cancel()
	return c.client.BalanceAt(_ctx, addr, nil)
}

// GetLatestStateByID TBD
func (c *Client) GetLatestStateByID(ctx context.Context, addr common.Address, id *big.Int) (abi.IStateStateInfo, error) {
	var (
		latestState abi.IStateStateInfo
		err         error
	)
	if err = c.Call(func(c *ethclient.Client) error {
		stateContact, err := abi.NewState(addr, c)
		if err != nil {
			return err
		}
		latestState, err = stateContact.GetStateInfoById(&bind.CallOpts{Context: ctx}, id)
		return err
	}); err != nil {
		return latestState, err
	}
	return latestState, nil
}

// CallAuth performs a Smart Contract method call that requires authorization.
// This call requires a valid account with Ether that can be spent during the
// call.
func (c *Client) CallAuth(ctx context.Context, gasLimit uint64, privateKey *ecdsa.PrivateKey, fn func(*ethclient.Client, *bind.TransactOpts) (*types.Transaction, error)) (*types.Transaction, error) {
	if privateKey == nil {
		return nil, ErrPrivateKeyNil
	}

	gasPrice, err := c.getGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gasPrice: %v", err)
	}
	log.Debug(ctx, "Transaction metadata", "gasPrice", gasPrice)

	cid, err := c.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chainID: %v", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, cid)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction signer: %v", err)
	}
	auth.Value = big.NewInt(0) // in wei
	if gasLimit == 0 {
		auth.GasLimit = uint64(c.Config.DefaultGasLimit) // in units
	} else {
		auth.GasLimit = gasLimit // in units
	}
	auth.GasPrice = gasPrice

	tx, err := fn(c.client, auth)
	if err != nil && strings.Contains(err.Error(), "transaction underpriced") {
		// TODO:
		// this is done in an attempt to solve issue with incorrect default gasPrice
		// from Polygon Mumbai testnet. This MUST be handled in a more general way
		// to support resending transaction that have failed because of network issues
		oldGasPrice := auth.GasPrice.Int64()
		auth.GasPrice = gasPrice.Mul(gasPrice, new(big.Int).SetInt64(transactionUnderpricedIncrement))
		log.Debug(ctx, "underpriced transaction has been resent",
			"old gasPrice", oldGasPrice,
			"new gasPrice", auth.GasPrice.Int64())
		tx, err = fn(c.client, auth)
	}
	if tx != nil {
		log.Debug(ctx, "Transaction", "tx", tx.Hash().Hex(), "nonce", tx.Nonce())
	}
	return tx, err
}

// ContractData eth smart-contract data
type ContractData struct {
	Address common.Address
	Tx      *types.Transaction
	Receipt *types.Receipt
}

// Call performs a read only Smart Contract method call.
func (c *Client) Call(fn func(*ethclient.Client) error) error {
	return fn(c.client)
}

func (c *Client) waitReceipt(ctx context.Context, txID common.Hash, timeout time.Duration) (*types.Receipt, error) {
	var err error
	var receipt *types.Receipt

	log.Debug(ctx, "Waiting for receipt", "tx", txID.Hex())

	start := time.Now()
	for {
		receipt, err = c.client.TransactionReceipt(ctx, txID)
		if err != nil {
			log.Debug(ctx, "get transaction receipt: ", "error", err)
		}

		if receipt != nil || time.Since(start) >= timeout {
			break
		}

		time.Sleep(c.Config.WaitReceiptCycleTime)
	}

	if receipt == nil {
		log.Debug(ctx, "Pending transaction / Wait receipt timeout", "tx", txID.Hex())
		return receipt, ErrReceiptNotReceived
	}
	log.Debug(ctx, "Receipt received", "tx", txID.Hex())

	return receipt, err
}

func (c *Client) waitBlock(ctx context.Context, timeout time.Duration, confirmationBlock *big.Int) error {
	var err error
	var blockNumber *big.Int

	start := time.Now()
	for {
		blockNumber, err = c.CurrentBlock(ctx)
		if err != nil {
			log.Error(ctx, "couldn't get the current block number", "err", err)
			break
		}
		if time.Since(start) >= timeout {
			err = errors.New("time out error during block number fetch")
			break
		}
		if blockNumber.Cmp(confirmationBlock) == 1 {
			break
		}

		time.Sleep(c.Config.WaitBlockCycleTime)
	}

	if err != nil {
		return err
	}

	if blockNumber == nil {
		return errors.New("couldn't fetch block number")
	}
	return nil
}

// CurrentBlock returns the current block number in the blockchain
func (c *Client) CurrentBlock(ctx context.Context) (*big.Int, error) {
	_ctx, cancel := context.WithTimeout(ctx, c.Config.RPCResponseTimeout)
	defer cancel()
	header, err := c.client.HeaderByNumber(_ctx, nil)
	if err != nil {
		return nil, err
	}
	return header.Number, nil
}

// ChainID get chain id.
func (c *Client) ChainID(ctx context.Context) (*big.Int, error) {
	_ctx, cancel := context.WithTimeout(ctx, c.Config.RPCResponseTimeout)
	defer cancel()
	cid, err := c.client.ChainID(_ctx)
	if err != nil {
		return nil, err
	}
	return cid, nil
}

// BlockByNumber get eth block by block number
func (c *Client) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	_ctx, cancel := context.WithTimeout(ctx, c.Config.RPCResponseTimeout)
	defer cancel()
	block, err := c.client.BlockByNumber(_ctx, number)
	if err != nil {
		return nil, err
	}
	return block, nil
}

// HeaderByNumber get eth block by block number
func (c *Client) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	_ctx, cancel := context.WithTimeout(ctx, c.Config.RPCResponseTimeout)
	defer cancel()
	header, err := c.client.HeaderByNumber(_ctx, number)
	if err != nil {
		return nil, err
	}
	return header, nil
}

// GetTransactionReceiptByID get tx receipt by tx id
func (c *Client) GetTransactionReceiptByID(ctx context.Context, txID string) (*types.Receipt, error) {
	_ctx, cancel := context.WithTimeout(ctx, c.Config.RPCResponseTimeout)
	defer cancel()
	receipt, err := c.client.TransactionReceipt(_ctx, common.HexToHash(txID))
	if err != nil {
		return nil, err
	}

	if receipt == nil {
		log.Debug(ctx, "Pending transaction", "tx", txID)
		return nil, ErrReceiptNotReceived
	}
	return receipt, nil
}

// WaitTransactionReceiptByID wait for transaction receipt
func (c *Client) WaitTransactionReceiptByID(ctx context.Context, txID string) (*types.Receipt, error) {
	return c.waitReceipt(ctx, common.HexToHash(txID), c.Config.ReceiptTimeout)
}

// WaitForBlock wait for eth block
func (c *Client) WaitForBlock(ctx context.Context, confirmationBlock *big.Int) error {
	return c.waitBlock(ctx, c.Config.ConfirmationTimeout, confirmationBlock)
}

// GetTransactionByID return the transaction by ID
func (c *Client) GetTransactionByID(ctx context.Context, txID string) (*types.Transaction, bool, error) {
	return c.client.TransactionByHash(ctx, common.HexToHash(txID))
}

// CreateTxOpts creates a new transaction signer
func (c *Client) CreateTxOpts(ctx context.Context, kmsKey kms.KeyID) (*bind.TransactOpts, error) {
	//nolint:all
	addr, err := c.getAddress(kmsKey)
	if err != nil {
		return nil, err
	}

	sigFn := c.signerFnFactory(ctx, kmsKey)

	opts := &bind.TransactOpts{
		From:   addr,
		Signer: sigFn,
	}

	if !c.Config.GasLess { // Some Ethereum nodes don't support eth_maxPriorityFeePerGas so we set GasLess = true
		tip, err := c.suggestGasTipCap(ctx)
		if err != nil {
			return nil, err
		}
		opts.GasPrice = nil
		gasLimit := uint64(0)
		opts.GasLimit = gasLimit
		opts.GasTipCap = tip
		return opts, nil
	}

	gasPrice, err := c.getGasPrice(ctx)
	if err != nil {
		log.Error(ctx, "failed to get gas price", "err", err)
		return nil, err
	}

	opts = &bind.TransactOpts{
		From:     addr,
		Signer:   sigFn,
		GasPrice: gasPrice,
		GasLimit: uint64(c.Config.DefaultGasLimit),
		Context:  ctx,
		NoSend:   false,
	}

	return opts, nil
}

// TransactionParams settings for transaction.
type TransactionParams struct {
	BaseFee     *big.Int
	GasTips     *big.Int
	Nonce       *uint64
	FromAddress common.Address
	ToAddress   common.Address
	Payload     []byte
	Value       *big.Int
}

// CreateRawTx raw transaction.
func (c *Client) CreateRawTx(ctx context.Context, txParams TransactionParams) (*types.Transaction, error) {
	if txParams.Nonce == nil {
		_ctx, cancel := context.WithTimeout(ctx, c.Config.RPCResponseTimeout)
		defer cancel()
		nonce, err := c.client.PendingNonceAt(_ctx, txParams.FromAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to get nonce: %v", err)
		}
		txParams.Nonce = &nonce
	}

	if txParams.Value == nil {
		txParams.Value = big.NewInt(0)
	}

	baseTx := &types.DynamicFeeTx{
		To:    &txParams.ToAddress,
		Nonce: *txParams.Nonce,
		Gas:   uint64(c.Config.DefaultGasLimit),
		Value: txParams.Value,
		Data:  txParams.Payload,
	}

	if !c.Config.GasLess {
		_ctx2, cancel2 := context.WithTimeout(ctx, c.Config.RPCResponseTimeout)
		defer cancel2()
		gasLimit, err := c.client.EstimateGas(_ctx2, ethereum.CallMsg{
			From:  txParams.FromAddress, // the sender of the 'transaction'
			To:    &txParams.ToAddress,
			Gas:   0,              // wei <-> gas exchange ratio
			Value: txParams.Value, // amount of wei sent along with the call
			Data:  txParams.Payload,
		})

		baseTx.Gas = gasLimit

		if err != nil {
			return nil, fmt.Errorf("failed to estimate gas: %v", err)
		}

		latestBlockHeader, err := c.HeaderByNumber(ctx, nil)
		if err != nil {
			return nil, err
		}

		if txParams.BaseFee == nil {
			// since ETH and Optimism blockchain already supports London fork.
			// no need set special block.
			baseFee := eip1559.CalcBaseFee(&params.ChainConfig{LondonBlock: big.NewInt(1)}, latestBlockHeader)

			// add 25% to baseFee. baseFee always small value.
			// since we use dynamic fee transactions we will get not used gas back.
			b := math.Round(float64(baseFee.Int64()) * feeIncrement)
			baseFee = big.NewInt(int64(b))
			txParams.BaseFee = baseFee
		}

		if txParams.GasTips == nil {
			_ctx3, cancel3 := context.WithTimeout(ctx, c.Config.RPCResponseTimeout)
			defer cancel3()
			gasTip, err := c.client.SuggestGasTipCap(_ctx3)
			// since hardhad doesn't support 'eth_maxPriorityFeePerGas' rpc call.
			// we should hardcode 0 as a mainer tips. More information: https://github.com/NomicFoundation/hardhat/issues/1664#issuecomment-1149006010
			if err != nil && strings.Contains(err.Error(), "eth_maxPriorityFeePerGas not found") {
				log.Error(ctx, "failed get suggest gas tip: %s. use 0 instead", "err", err)
				gasTip = big.NewInt(0)
			} else if err != nil {
				return nil, fmt.Errorf("failed get suggest gas tip: %v", err)
			}
			txParams.GasTips = gasTip
		}

		maxGasPricePerFee := big.NewInt(0).Add(txParams.BaseFee, txParams.GasTips)

		baseTx.GasTipCap = txParams.GasTips
		baseTx.GasFeeCap = maxGasPricePerFee

	}

	tx := types.NewTx(baseTx)

	return tx, nil
}

// SendRawTx send raw transaction.
func (c *Client) SendRawTx(ctx context.Context, tx *types.Transaction) error {
	_ctx, cancel := context.WithTimeout(ctx, c.Config.RPCResponseTimeout)
	defer cancel()
	return c.client.SendTransaction(_ctx, tx)
}

// getGasPrice returns suggested gas price within configured bounds
func (c *Client) getGasPrice(ctx context.Context) (*big.Int, error) {
	gasPrice := new(big.Int)
	zero := big.NewInt(0)

	// if configured min gas price == max gas price and is not zero, then force this value
	if c.Config.MinGasPrice != nil && c.Config.MinGasPrice.Cmp(zero) == Gt &&
		c.Config.MaxGasPrice != nil && c.Config.MinGasPrice.Cmp(c.Config.MaxGasPrice) == Eq {
		return gasPrice.Set(c.Config.MaxGasPrice), nil
	}

	_ctx, cancel := context.WithTimeout(ctx, c.Config.RPCResponseTimeout)
	defer cancel()
	suggestedGasPrice, err := c.client.SuggestGasPrice(_ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get suggested gas price: %v", err)
	}

	// increase suggested gas price by 10% for better confirmation speed
	inc := new(big.Int).Set(suggestedGasPrice)
	inc.Div(inc, new(big.Int).SetUint64(gasPriceIncrement))
	suggestedGasPrice.Add(suggestedGasPrice, inc)

	gasPrice.Set(suggestedGasPrice)

	// correct value if estimated gas price is less than configured min value
	if c.Config.MinGasPrice != nil && c.Config.MinGasPrice.Cmp(zero) == Gt &&
		gasPrice.Cmp(c.Config.MinGasPrice) == Lt {
		gasPrice.Set(c.Config.MinGasPrice)
	}
	// correct value if estimated gas price is more than configured max value
	if c.Config.MaxGasPrice != nil && c.Config.MaxGasPrice.Cmp(zero) == Gt &&
		gasPrice.Cmp(c.Config.MaxGasPrice) == Gt {
		gasPrice.Set(c.Config.MaxGasPrice)
	}

	if gasPrice.Cmp(suggestedGasPrice) != Eq {
		log.Debug(ctx, "Transaction metadata",
			"suggested gas price", suggestedGasPrice,
			"corrected gas price", gasPrice)
	}

	return gasPrice, err
}

// getAddress - get address by keyID
func (c *Client) getAddress(k kms.KeyID) (common.Address, error) {
	if c.kms == nil {
		return common.Address{}, errors.Join(errors.New("the signer is read-only"))
	}
	bytesPubKey, err := c.kms.PublicKey(k)
	if err != nil {
		return common.Address{}, err
	}
	var pubKey *ecdsa.PublicKey
	switch len(bytesPubKey) {
	case CompressedPublicKeyLength:
		pubKey, err = crypto.DecompressPubkey(bytesPubKey)
	case AwsKmsPublicKeyLength:
		pubKey, err = kms.DecodeAWSETHPubKey(context.Background(), bytesPubKey)
		if err != nil {
			return common.Address{}, err
		}
	default:
		pubKey, err = crypto.UnmarshalPubkey(bytesPubKey)
	}
	if err != nil {
		return common.Address{}, err
	}
	fromAddress := crypto.PubkeyToAddress(*pubKey)
	return fromAddress, nil
}

func (c *Client) signerFnFactory(ctx context.Context, signingKeyID kms.KeyID) func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
	return func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
		if c.kms == nil {
			return nil, errors.Join(errors.New("the signer is read-only"))
		}

		ch, err := c.ChainID(ctx)
		if err != nil {
			return nil, err
		}

		signer := types.LatestSignerForChainID(ch)
		h := signer.Hash(tx)

		sig, err := c.kms.Sign(ctx, signingKeyID, h[:])
		if err != nil {
			return nil, err
		}

		return tx.WithSignature(signer, sig)
	}
}

func (c *Client) suggestGasTipCap(ctx context.Context) (*big.Int, error) {
	ctxWT, cancel3 := context.WithTimeout(ctx, c.Config.RPCResponseTimeout)
	defer cancel3()

	tip, err := c.client.SuggestGasTipCap(ctxWT)
	// since hardhat doesn't support 'eth_maxPriorityFeePerGas' rpc call.
	// we should hard code 0 as a mainer tips. More information: https://github.com/NomicFoundation/hardhat/issues/1664#issuecomment-1149006010
	if err != nil && strings.Contains(err.Error(), "eth_maxPriorityFeePerGas not found") {
		log.Info(ctx, "failed get suggest gas tip: %s. use 0 instead", "err", err)
		tip = big.NewInt(0)
	} else if err != nil {
		return nil, errors.Join(err, errors.New("failed get suggest gas tip"))
	}

	return tip, nil
}
