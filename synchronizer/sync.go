package synchronizer

import (
	"context"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/cool-develope/dex-router/etherman"
)

// Synchronizer interface
type Synchronizer interface {
	Sync() error
	Stop()
}

// EtherMan inteface
type EtherMan interface {
	GetRates() (map[etherman.Pair]float64, error)
}

// Symbols is an order of symbols.
type Symbols []string

// Quote is a struct represent the best combination of pair.
type Quote struct {
	symbols Symbols
	rate    float64
}

// ClientSynchronizer connects network.
type ClientSynchronizer struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	etherMan   EtherMan
	rates      map[etherman.Pair]float64
	quotes     map[etherman.Pair]Quote
	rwMutex    sync.RWMutex
}

var waitDuration = time.Duration(1000000000) // 1 second
var updateLimit = 3

// NewClientSynchronizer creates new synchronizer.
func NewClientSynchronizer(etherMan EtherMan) *ClientSynchronizer {
	ctx, cancel := context.WithCancel(context.Background())
	return &ClientSynchronizer{
		ctx:        ctx,
		cancelFunc: cancel,
		etherMan:   etherMan,
		rates:      make(map[etherman.Pair]float64),
		quotes:     make(map[etherman.Pair]Quote),
	}
}

// Sync starts the synchronizer.
func (s *ClientSynchronizer) Sync() error {
	log.Default().Println("Synchronization started!")

	for {
		select {
		case <-s.ctx.Done():
			log.Default().Println("synchronizer ctx done.")
			return nil
		case <-time.After(waitDuration):
			pairRates, err := s.etherMan.GetRates()
			if err != nil {
				log.Default().Printf("synchronizer sync rate failed: %v", err)
				if err != etherman.ErrNotRegisteredPair {
					continue
				}
				return err
			}
			err = s.updateRates(pairRates)
			if err != nil {
				return err
			}
		}
	}
}

// Stop function stops the synchronizer.
func (s *ClientSynchronizer) Stop() {
	s.cancelFunc()
}

func (s *ClientSynchronizer) updateRates(pairRates map[etherman.Pair]float64) error {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()

	s.rates = pairRates
	s.quotes = make(map[etherman.Pair]Quote)
	edges := make(map[etherman.Pair]float64)

	for key, value := range pairRates {
		if value > 0 {
			logRate := math.Log(value)
			edges[key] = logRate
			s.quotes[key] = Quote{
				symbols: Symbols{key.TokenA, key.TokenB},
				rate:    logRate,
			}
			s.quotes[key] = Quote{
				symbols: Symbols{key.TokenB, key.TokenA},
				rate:    -logRate,
			}
		} else {
			log.Default().Printf("negative rate of pair: %v", key)
		}

	}

	for i := 0; i < updateLimit; i++ {
		cons := make(map[string][]Quote)
		for key, quote := range s.quotes {
			quotes, found := cons[key.TokenA]
			if !found {
				quotes = make([]Quote, 0)

			}
			quotes = append(quotes, quote)
			cons[key.TokenA] = quotes
		}

		for pair, price := range edges {
			source := pair.TokenB
			for _, quote := range cons[source] {
				visited := false
				for _, symbol := range quote.symbols {
					if symbol == pair.TokenA {
						visited = true
						break
					}
				}
				if !visited {
					nPair := etherman.Pair{
						TokenA: pair.TokenA,
						TokenB: quote.symbols[len(quote.symbols)-1],
					}
					nQuote, found := s.quotes[nPair]
					if !found || (found && price+quote.rate < nQuote.rate) {
						symbols := []string{pair.TokenA}
						symbols = append(symbols, quote.symbols...)
						s.quotes[nPair] = Quote{
							rate:    price + quote.rate,
							symbols: symbols,
						}
					}
				}
			}
		}
	}

	for key, quote := range s.quotes {
		quote.rate = math.Exp(quote.rate)
		s.quotes[key] = quote
	}

	return nil
}

func (s *ClientSynchronizer) GetRate(pair etherman.Pair) (float64, error) {
	s.rwMutex.RLock()
	defer s.rwMutex.RUnlock()
	rate, found := s.rates[pair]
	if !found {
		return 0, fmt.Errorf("not found pair: %v", pair)
	}

	return rate, nil
}

func (s *ClientSynchronizer) GetQuote(pair etherman.Pair) (*Quote, error) {
	s.rwMutex.RLock()
	defer s.rwMutex.RUnlock()
	quote, found := s.quotes[pair]
	if !found {
		return nil, fmt.Errorf("not found pair: %v", pair)
	}

	return &quote, nil
}
