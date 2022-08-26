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

const eps = 1e-10

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
	Symbols Symbols
	Rate    float64
}

type neighbour struct {
	token string
	rate  float64
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
	cons := make(map[string][]neighbour)

	for key, value := range pairRates {
		if value > 0 {
			logRate := math.Log(value)

			// build neightbours
			neis, found := cons[key.TokenA]
			if !found {
				neis = make([]neighbour, 0)

			}
			neis = append(neis, neighbour{
				token: key.TokenB,
				rate:  logRate,
			})
			cons[key.TokenA] = neis
			// reverse connection
			neis, found = cons[key.TokenB]
			if !found {
				neis = make([]neighbour, 0)

			}
			neis = append(neis, neighbour{
				token: key.TokenA,
				rate:  -logRate,
			})
			cons[key.TokenB] = neis
		} else {
			log.Default().Printf("negative rate of pair: %v", key)
		}

		// build the initial quotes
		pairA, pairB := etherman.Pair{TokenA: key.TokenA, TokenB: key.TokenA}, etherman.Pair{TokenA: key.TokenB, TokenB: key.TokenB}
		s.quotes[pairA] = Quote{
			Symbols: Symbols{key.TokenA},
			Rate:    0,
		}
		s.quotes[pairB] = Quote{
			Symbols: Symbols{key.TokenB},
			Rate:    0,
		}
	}
	// repeats Bellman-Ford algorithm due to negative cycle
	for i := 0; i < updateLimit; i++ {
		for pair, quote := range s.quotes {
			source := pair.TokenB
			// not update the cycle
			if source == pair.TokenA && len(quote.Symbols) > 1 {
				continue
			}
			for _, nei := range cons[source] {
				// to check the cycle in the path
				visited := false
				for i, symbol := range quote.Symbols {
					// accept the simple cycle
					if i > 0 && symbol == nei.token {
						visited = true
						break
					}
				}
				if !visited {
					nPair := etherman.Pair{
						TokenA: pair.TokenA,
						TokenB: nei.token,
					}
					nQuote, found := s.quotes[nPair]
					if !found || (found && quote.Rate+nei.rate > nQuote.Rate+eps) {
						s.quotes[nPair] = Quote{
							Rate:    quote.Rate + nei.rate,
							Symbols: append(append([]string(nil), quote.Symbols...), nei.token),
						}
					}
				}
			}
		}
	}

	for key, quote := range s.quotes {
		quote.Rate = math.Exp(quote.Rate)
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
