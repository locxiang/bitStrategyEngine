/**
购买策略

设置止损止盈策略
 */
package events

import (
	"github.com/locxiang/bitStrategyEngine/dispatcher"
)

type Purchase struct {
	symbol    string
	price     float64 //购买价
	goodPrice float64 //获利抛售
	badPrice  float64 //止损抛售
	orderID   int64   //订单id
}

func (p *Purchase) Handle(strategy dispatcher.Strategy) {
	panic("implement me")
}

func (p *Purchase) buy() bool {
	return true
}

func (p *Purchase) sell() bool {
	return true
}