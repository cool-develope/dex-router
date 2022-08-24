package etherman

import (
	"errors"
	"math/big"

	"github.com/cool-develope/dex-router/etherman/uniswap/ERC20"
	"github.com/cool-develope/dex-router/etherman/uniswap/factory"
	"github.com/cool-develope/dex-router/etherman/uniswap/pair"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	// ErrStorageNotFound is used when the pair is not registered.
	ErrNotRegisteredPair = errors.New("not registered pair")
)

type ethClienter interface {
	ethereum.ChainReader
	ethereum.LogFilterer
	ethereum.TransactionReader
	ethereum.ContractCaller
}

// Pair is a pair of tokens.
type Pair struct {
	TokenA, TokenB string
}

// Client is a simple implementation of EtherMan.
type Client struct {
	EtherClient ethClienter
	Factory     *factory.Factory
	Pairs       map[Pair]*pair.Pair
}

// NewClient create a new client.
func NewClient(url string, factoryAddr common.Address) (*Client, error) {
	ethClient, err := ethclient.Dial(url)
	if err != nil {
		return nil, err
	}

	factorySC, err := factory.NewFactory(factoryAddr, ethClient)
	if err != nil {
		return nil, err
	}

	client := &Client{
		EtherClient: ethClient,
		Factory:     factorySC,
		Pairs:       make(map[Pair]*pair.Pair),
	}

	return client, nil
}

func (etherMan *Client) getTokenSymbol(tokenAddr common.Address) (string, error) {
	tokenSC, err := ERC20.NewERC20(tokenAddr, etherMan.EtherClient.(*ethclient.Client))
	if err != nil {
		return "", err
	}
	return tokenSC.Symbol(nil)
}

// GetPair returns the dedicated pair.
func (etherMan *Client) RegisterPair(tokenA, tokenB common.Address) (*Pair, error) {
	pairAddr, err := etherMan.Factory.GetPair(nil, tokenA, tokenB)
	if err != nil {
		return nil, err
	}
	pairSC, err := pair.NewPair(pairAddr, etherMan.EtherClient.(*ethclient.Client))
	if err != nil {
		return nil, err
	}

	tokenAName, err := etherMan.getTokenSymbol(tokenA)
	if err != nil {
		return nil, err
	}
	tokenBName, err := etherMan.getTokenSymbol(tokenB)
	if err != nil {
		return nil, err
	}

	pair := Pair{
		TokenA: tokenAName,
		TokenB: tokenBName,
	}
	etherMan.Pairs[pair] = pairSC

	return &pair, nil
}

func (etherMan *Client) getPairRate(pair Pair) (float64, error) {
	pairSC, found := etherMan.Pairs[pair]
	if !found {
		return 0, ErrNotRegisteredPair
	}

	reserve, err := pairSC.GetReserves(nil)
	if err != nil {
		return 0, err
	}

	reserveA, reserveB := new(big.Float).SetInt(reserve.Reserve0), new(big.Float).SetInt(reserve.Reserve1)
	rate, _ := new(big.Float).Quo(reserveA, reserveB).Float64()
	return rate, nil
}

// GetRates returns rates for all pairs.
func (etherMan *Client) GetRates() (map[Pair]float64, error) {
	rates := make(map[Pair]float64)

	for key := range etherMan.Pairs {
		rate, err := etherMan.getPairRate(key)
		if err != nil {
			return rates, err
		}
		rates[key] = rate
	}

	return rates, nil
}
