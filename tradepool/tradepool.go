package tradepool

import (
	"github.com/binance-exchange/go-binance"
	"time"
	"fmt"
	"github.com/locxiang/bitStrategyEngine/dispatcher"
	"sync"
)

/**
交易数据池
此数据池实现三种能力
1. 交易明细
2. 交易量统计
3. 价格波动比率
 */
type TradePool struct {
	Duration      time.Duration //在一定时间内
	Symbol        string        //类别
	aggTrades     []*AggTrade
	sellCount     float64 //卖出量
	buyCount      float64 //买入量
	strategyGroup []dispatcher.Strategy
	sync.Mutex
}

type AggTrade struct {
	*binance.AggTrade
	Direction binance.OrderSide //方向  Buy  Sell
}

//线程池种类的唯一id   同一个策略的数据池只有一个id
func (e *TradePool) Key() string {
	key := e.Symbol + e.Duration.String()
	return key
}

func (e *TradePool) AcceptAggTrade(trade *binance.AggTrade) {
	defer e.Unlock()
	e.Lock()

	fmt.Printf("\n\n\n\n收到一条数据：%+v \n", trade)
	e.addAggTrade(trade)
	e.removeExpiredTrade()
}

//增加策略
func (e *TradePool) RegisterStrategy(s dispatcher.Strategy) {
	e.strategyGroup = append(e.strategyGroup, s)
}

//注销策略
func (e *TradePool) UnregisterStrategy(s dispatcher.Strategy) {
	for i, v := range e.strategyGroup {
		if v == s {
			e.strategyGroup = append(e.strategyGroup[:i], e.strategyGroup[i+1:]...)
			break
		}
	}
}

func (e *TradePool) StrategyAll() []dispatcher.Strategy {
	return e.strategyGroup
}

func (e *TradePool) GetSymbol() string {
	return e.Symbol
}

func (e *TradePool) GetDuration() time.Duration {
	return e.Duration
}

//最初价格
func (e *TradePool) FirstPrice() (float64) {
	return e.aggTrades[0].Price
}

//最新价格
func (e *TradePool) LastPrice() (float64) {
	l := len(e.aggTrades)
	if l == 0 {
		return -1
	}
	return e.aggTrades[l-1].Price
}

func (e *TradePool) addAggTrade(trade *binance.AggTrade) {

	tr := &AggTrade{
		AggTrade: trade,
	}
	//TODO  此处的订单是买还是 卖 计算有误，需要根据挂单数据做匹配
	if trade.Price > e.LastPrice() {
		tr.Direction = binance.SideBuy
		e.buyCount += tr.Quantity
	} else {
		tr.Direction = binance.SideSell
		e.sellCount += tr.Quantity
	}
	e.aggTrades = append(e.aggTrades, tr)

	fmt.Printf("当前价格;%f, 最新价格：%f  波动率：%f%% \n", trade.Price, e.LastPrice(), e.Ratio()*100)
	//for _, s := range e.aggTrades {
	//	fmt.Printf("看看e.s.Direction: %s\t %f,%f \n", s.Direction, s.Quantity, s.Price)
	//}

}

//过滤掉多余数据
func (e *TradePool) removeExpiredTrade() {
	t1 := time.Now()
	for {
		t2 := e.aggTrades[0].Timestamp.Add(e.Duration)
		if t1.After(t2) {
			if e.aggTrades[0].Direction == binance.SideBuy {
				e.buyCount -= e.aggTrades[0].Quantity
			} else {
				e.sellCount -= e.aggTrades[0].Quantity
			}
			e.aggTrades = e.aggTrades[1:]

			continue
		}
		break
	}
}

// 获取此数据池的百分率
func (e *TradePool) Ratio() float64 {
	ratio := (e.LastPrice() - e.FirstPrice()) / e.FirstPrice()
	return ratio
}

func (e *TradePool) Close() error {
	panic("implement me")
}
