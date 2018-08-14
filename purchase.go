/**
购买策略

设置止损止盈策略
 */
package main

import "github.com/binance-exchange/go-binance"

type Purchase struct {
	symbol    string
	price     float64 //购买价
	goodPrice float64 //获利抛售
	badPrice  float64 //止损抛售
	orderID   int64   //订单id
}

func (p *Purchase) Buy() bool {
	return true
}

func (p *Purchase) Sell() bool {
	return true
}

func (p *Purchase) AcceptAggTrade(trade *binance.AggTrade) {

	if trade.Price >= p.goodPrice {

	}
}

func (p *Purchase) Symbol() string {
	panic("implement me")
}

func (p *Purchase) Close() error {
	panic("implement me")
}
