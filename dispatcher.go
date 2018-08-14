package main

import (
	"github.com/binance-exchange/go-binance"
)

type Consumer interface {
	AcceptAggTrade(trade *binance.AggTrade)
	Symbol() string
	Close() error
}

type dispatcher struct {
	// Registered Consumer.
	Consumers     map[Consumer]bool
	EventAggTrade chan *binance.AggTradeEvent
	register      chan Consumer
	unregister    chan Consumer
}

func NewDispatcher() *dispatcher {
	return &dispatcher{
		EventAggTrade: make(chan *binance.AggTradeEvent),
		register:      make(chan Consumer),
		unregister:    make(chan Consumer),
		Consumers:     make(map[Consumer]bool),
	}
}

func (d *dispatcher) run() {
	for {
		select {
		case c := <-d.register:
			d.Consumers[c] = true
		case c := <-d.unregister:
			if _, ok := d.Consumers[c]; ok {
				delete(d.Consumers, c)
				c.Close()
			}
		case m := <-d.EventAggTrade:
			for c := range d.Consumers {

				if c.Symbol() == m.Symbol {
					c.AcceptAggTrade(&m.AggTrade)
				}
			}
		}
	}
}
