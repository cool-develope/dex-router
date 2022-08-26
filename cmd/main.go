package main

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"

	"github.com/cool-develope/dex-router/etherman"
	"github.com/cool-develope/dex-router/server"
	"github.com/cool-develope/dex-router/synchronizer"
	"github.com/ethereum/go-ethereum/common"
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
		tokenA: WETH,
		tokenB: USDC,
	},
	{
		tokenA: WETH,
		tokenB: WISE,
	},
	{
		tokenA: USDC,
		tokenB: DAI,
	},
	{
		tokenA: WETH,
		tokenB: DAI,
	},
	{
		tokenA: USDC,
		tokenB: USDT,
	},
	{
		tokenA: USDT,
		tokenB: WETH,
	},
}

func main() {
	etherMan, err := etherman.NewClient(clientURL+os.Getenv("INFURA_KEY"), common.HexToAddress(factoryAddr))
	if err != nil {
		log.Default().Panic(err)
	}

	for _, pair := range pairs {
		_, err := etherMan.RegisterPair(common.HexToAddress(pair.tokenA), common.HexToAddress(pair.tokenB))
		if err != nil {
			log.Default().Panic(err)
		}
	}

	sync := synchronizer.NewClientSynchronizer(etherMan)
	go sync.Sync()

	handler := server.NewHandler(sync)

	err = rpc.Register(handler)
	if err != nil {
		log.Default().Panic(err)
	}

	rpc.HandleHTTP()
	//start listening for messages on port 1234
	l, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatalf("Couldn't start listening on port 1234. Error %s", e)
	}
	log.Println("Serving RPC handler")
	err = http.Serve(l, nil)
	if err != nil {
		log.Fatalf("Error serving: %s", err)
	}

}
