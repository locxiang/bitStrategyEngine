package dispatcher

import (
	"github.com/binance-exchange/go-binance"
	"time"
	"sync"
)

//事件接口
type EventFace interface {
	Handle(strategy Strategy)
}

type Strategy interface {
	Check(p TradePool) bool //检测策略是否命中
	HandleEvents()          //如果命中策略就执行的方法
	Valid() bool            //判断此策略是否开启
	GetPool() TradePool		//返回线程池
}

type TradePool interface {
	AcceptAggTrade(trade *binance.AggTrade)
	GetSymbol() string
	Key() string
	Close() error
	GetDuration() time.Duration
	FirstPrice() float64
	LastPrice() float64
	RegisterStrategy(s Strategy)
	UnregisterStrategy(s Strategy)
	StrategyAll() []Strategy
	Ratio() float64
}

var tradePools struct {
	m map[string]TradePool
	sync.Mutex
}

var Disp *Dispatcher

func init() {
	tradePools.m = make(map[string]TradePool)
	Disp = NewDispatcher()
	go Disp.Run()
}

//注册数据池
func RegisterPools(t TradePool) {
	defer tradePools.Unlock()
	tradePools.Lock()

	if _, ok := tradePools.m[t.Key()]; !ok {
		tradePools.m[t.Key()] = t
		Disp.Register <- t
	}

}

type Dispatcher struct {
	// Registered TradePool.
	Pools         map[TradePool]bool
	EventAggTrade chan *binance.AggTradeEvent
	Register      chan TradePool
	Unregister    chan TradePool
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		EventAggTrade: make(chan *binance.AggTradeEvent),
		Register:      make(chan TradePool),
		Unregister:    make(chan TradePool),
		Pools:         make(map[TradePool]bool),
	}
}

func (d *Dispatcher) Run() {
	for {
		select {
		case pool := <-d.Register:
			d.Pools[pool] = true
		case pool := <-d.Unregister:
			if _, ok := d.Pools[pool]; ok {
				delete(d.Pools, pool)
				pool.Close()
			}
		case m := <-d.EventAggTrade:
			for pool := range d.Pools {
				if pool.GetSymbol() == m.Symbol {
					pool.AcceptAggTrade(&m.AggTrade)
					// 循环检查策略
					for _, s := range pool.StrategyAll() {
						checkPoolStrategy(s, pool)
					}

				}
			}
		}
	}
}

//循环检查策略
func checkPoolStrategy(s Strategy, pool TradePool) bool {
	//策略有效 并且命中  执行headle
	if s.Valid() && s.Check(pool) {
		go s.HandleEvents() //避免执行事件太长耽误其他策略
		return true
	}

	return false

}
