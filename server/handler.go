package server

import (
	"github.com/cool-develope/dex-router/etherman"
	"github.com/cool-develope/dex-router/synchronizer"
)

// Synchronizer interface
type Synchronizer interface {
	GetRate(pair etherman.Pair) (float64, error)
	GetQuote(pair etherman.Pair) (*synchronizer.Quote, error)
}

// Handler is a rpc handler.
type Handler struct {
	sync Synchronizer
}

func NewHandler(sync Synchronizer) *Handler {
	return &Handler{
		sync: sync,
	}
}

// GetRate returns the rate of the dedicated pair.
func (rh *Handler) GetRate(pair etherman.Pair, rate *float64) error {
	r, err := rh.sync.GetRate(pair)
	*rate = r
	return err
}

// GetQuote returns the quote of the dedicated pair.
func (rh *Handler) GetQuote(pair etherman.Pair, quote *synchronizer.Quote) error {
	q, err := rh.sync.GetQuote(pair)
	*quote = *q
	return err
}
