package etherman_test

import (
	"os"
	"testing"

	"github.com/cool-develope/dex-router/etherman"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

const (
	clientURL   = "https://mainnet.infura.io/v3/"
	factoryAddr = "0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f"

	WETH = "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"
	WISE = "0x66a0f676479cee1d7373f3dc2e2952778bff5bd6"
	USDC = "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"
	USDT = "0xdac17f958d2ee523a2206206994597c13d831ec7"
	DAI  = "0x6b175474e89094c44da98b954eedeac495271d0f"
)

var pairs = []struct {
	tokenA string
	tokenB string
}{
	{
		tokenA: USDC,
		tokenB: WETH,
	},
	{
		tokenA: WISE,
		tokenB: WETH,
	},
	{
		tokenA: DAI,
		tokenB: USDC,
	},
	{
		tokenA: DAI,
		tokenB: WETH,
	},
	{
		tokenA: USDC,
		tokenB: USDT,
	},
	{
		tokenA: WETH,
		tokenB: USDT,
	},
}

func TestEtherClient(t *testing.T) {
	etherMan, err := etherman.NewClient(clientURL+os.Getenv("INFURA_KEY"), common.HexToAddress(factoryAddr))
	require.NoError(t, err)

	for _, pair := range pairs {
		_, err := etherMan.RegisterPair(common.HexToAddress(pair.tokenA), common.HexToAddress(pair.tokenB))
		require.NoError(t, err)
	}

	rates, err := etherMan.GetRates()
	require.NoError(t, err)
	require.Equal(t, len(pairs), len(rates))
	t.Logf("rates : %v", rates)
}
