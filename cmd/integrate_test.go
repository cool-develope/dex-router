//go:build e2e
// +build e2e

package main_test

import (
	"net/rpc"
	"testing"

	"github.com/cool-develope/dex-router/etherman"
	"github.com/cool-develope/dex-router/synchronizer"
	"github.com/stretchr/testify/require"
)

func TestMain(t *testing.T) {
	client, err := rpc.DialHTTP("tcp", ":1234")
	require.NoError(t, err)

	args := etherman.Pair{
		TokenA: "USDC",
		TokenB: "USDT",
	}

	var (
		resRate  float64
		resQuote synchronizer.Quote
	)
	err = client.Call("Handler.GetRate", args, &resRate)
	require.NoError(t, err)
	t.Logf("rate : %f", resRate)

	args = etherman.Pair{
		TokenA: "USDT",
		TokenB: "USDC",
	}
	err = client.Call("Handler.GetQuote", args, &resQuote)
	require.NoError(t, err)
	t.Logf("result : %v", resQuote)

	args = etherman.Pair{
		TokenA: "DAI",
		TokenB: "DAI",
	}
	err = client.Call("Handler.GetQuote", args, &resQuote)
	require.NoError(t, err)
	t.Logf("result : %v", resQuote)

	// require.True(t, false)
}
