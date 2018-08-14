package main

import (
	"github.com/binance-exchange/go-binance"
	"time"
	"fmt"
	"sync"
)

type Strategy interface {
	Check() bool         //检测策略是否命中
	Handle()             //如果命中策略就执行的方法
	TimeOut() *time.Time //失效时间
	Valid() bool         //检测策略状态是否有效
}

//交易数据池
type tradePool struct {
	duration      time.Duration //在一定时间内
	symbol        string        //类别
	aggTrades     []*AggTrade
	strategyGroup []Strategy
}

var tradePools sync.Map //所有数据池

type AggTrade struct {
	*binance.AggTrade
	Direction string //方向  Buy  Sell
}

//获取一个数据池
func GetPool(symbol string, duration time.Duration) *tradePool {
	e := &tradePool{
		duration:
		duration,
		symbol: symbol,
	}

	tp, _ := tradePools.LoadOrStore(e.Key(), e)
	return tp.(*tradePool)

	fmt.Printf("创建数据池[%s]：%+v \n", e.Key(), e)
	return e
}

func (e *tradePool) Key() string {
	key := e.symbol + e.duration.String()
	return key
}

func (e *tradePool) AcceptAggTrade(trade *binance.AggTrade) {
	fmt.Printf("\n\n\n\n收到一条数据：%+v \n", trade)
	e.add(trade)

	// 循环检查策略
	for _, s := range e.strategyGroup {
		//策略有效 并且命中  执行headle
		if s.Valid() && s.Check() {
			s.Handle()
		}
	}

}

//增加策略
func (e *tradePool) AddStrategy(s Strategy) {
	e.strategyGroup = append(e.strategyGroup, s)
}

func (e *tradePool) Symbol() string {
	return e.symbol
}

//最初价格
func (e *tradePool) firstPrice() (float64, error) {
	l := len(e.aggTrades)
	if l == 0 {
		return 0, fmt.Errorf("数据池为空")
	}
	return e.aggTrades[0].Price, nil
}

//最新价格
func (e *tradePool) lastPrice() (float64, error) {
	l := len(e.aggTrades)
	if l == 0 {
		return 0, fmt.Errorf("数据池为空")
	}
	return e.aggTrades[l-1].Price, nil
}

func (e *tradePool) add(trade *binance.AggTrade) {

	tr := &AggTrade{
		AggTrade: trade,
	}

	lastPrice, err := e.lastPrice()
	if err != nil {
		tr.Direction = "Buy"
	} else {
		if trade.Price > lastPrice {
			tr.Direction = "Buy"
		} else {
			tr.Direction = "Sell"
		}
	}
	e.aggTrades = append(e.aggTrades, tr)
	//过滤掉多余数据
	e.removeExpiredTrade()
}

func (e *tradePool) removeExpiredTrade() {
	t1 := time.Now()
	for {
		t2 := e.aggTrades[0].Timestamp.Add(e.duration)
		if t1.After(t2) {
			e.aggTrades = e.aggTrades[1:]
			continue
		}
		break
	}
}

// 获取此数据池的百分率
func (e *tradePool) Ratio() float64 {
	firstPrice, err := e.firstPrice()
	if err != nil {
		return 0
	}
	lastPrice, err := e.lastPrice()
	if err != nil {
		return 0
	}

	ratio := (lastPrice - firstPrice) / firstPrice

	return ratio
}

func (e *tradePool) Close() error {
	panic("implement me")
}
