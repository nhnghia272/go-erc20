package goerc20

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Config struct {
	ChainRpc   string
	PrivateKey string
	Contract   string
	Decimals   uint
}

func New(config Config) *Token {
	if !common.IsHexAddress(config.Contract) {
		panic(errors.New("invalid contract address"))
	}
	contractAbi, err := abi.JSON(strings.NewReader(ABI))
	if err != nil {
		panic(err)
	}
	client, err := ethclient.Dial(config.ChainRpc)
	if err != nil {
		panic(err)
	}
	privateKey, err := crypto.HexToECDSA(config.PrivateKey)
	if err != nil {
		panic(err)
	}
	chainId, err := client.NetworkID(context.Background())
	if err != nil {
		panic(err)
	}
	return &Token{
		client:          client,
		privateKey:      privateKey,
		chainId:         chainId,
		decimals:        big.NewInt(int64(math.Pow10(int(config.Decimals)))),
		contractAddress: common.HexToAddress(config.Contract),
		contractAbi:     contractAbi,
	}
}

type Token struct {
	client          *ethclient.Client
	privateKey      *ecdsa.PrivateKey
	chainId         *big.Int
	decimals        *big.Int
	contractAddress common.Address
	contractAbi     abi.ABI
}

func (s Token) privateKeyToAddress() common.Address {
	publicKey := s.privateKey.Public()
	publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
	return crypto.PubkeyToAddress(*publicKeyECDSA)
}

func (s Token) floatToBigInt(val float64, decimals *big.Int) *big.Int {
	bigval := new(big.Float)
	bigval.SetFloat64(val)

	coin := new(big.Float)
	coin.SetInt(decimals)

	bigval.Mul(bigval, coin)

	result := new(big.Int)
	bigval.Int(result)

	return result
}

func (s Token) SendTo(to string, val float64) error {
	if !common.IsHexAddress(to) {
		return errors.New("invalid to address")
	}
	if val <= 0 {
		return errors.New("value must be positive")
	}

	nonce, err := s.client.PendingNonceAt(context.Background(), s.privateKeyToAddress())
	if err != nil {
		return err
	}
	gasPrice, err := s.client.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}

	gasLimit := uint64(21000)
	amount := s.floatToBigInt(val, big.NewInt(int64(math.Pow10(18))))

	tx := types.NewTransaction(nonce, common.HexToAddress(to), amount, gasLimit, gasPrice, nil)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(s.chainId), s.privateKey)
	if err != nil {
		return err
	}

	return s.client.SendTransaction(context.Background(), signedTx)
}

func (s Token) Erc20SendTo(to string, val float64) error {
	if !common.IsHexAddress(to) {
		return errors.New("invalid to address")
	}
	if val <= 0 {
		return errors.New("value must be positive")
	}

	data, err := s.contractAbi.Pack("transfer", common.HexToAddress(to), s.floatToBigInt(val, s.decimals))
	if err != nil {
		return err
	}

	return s.writeContract(data)
}

func (s Token) Erc20SendFrom(from, to string, val float64) error {
	if !common.IsHexAddress(from) {
		return errors.New("invalid from address")
	}
	if !common.IsHexAddress(to) {
		return errors.New("invalid to address")
	}
	if val <= 0 {
		return errors.New("value must be positive")
	}

	data, err := s.contractAbi.Pack("transferFrom", common.HexToAddress(from), common.HexToAddress(to), s.floatToBigInt(val, s.decimals))
	if err != nil {
		return err
	}

	return s.writeContract(data)
}

func (s Token) writeContract(data []byte) error {
	nonce, err := s.client.PendingNonceAt(context.Background(), s.privateKeyToAddress())
	if err != nil {
		return err
	}
	gasPrice, err := s.client.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}
	gasLimit, err := s.client.EstimateGas(context.Background(), ethereum.CallMsg{From: s.privateKeyToAddress(), To: &s.contractAddress, Data: data})
	if err != nil {
		return err
	}

	tx := types.NewTransaction(nonce, s.contractAddress, big.NewInt(0), gasLimit, gasPrice, data)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(s.chainId), s.privateKey)
	if err != nil {
		return err
	}

	return s.client.SendTransaction(context.Background(), signedTx)
}
