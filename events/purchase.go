/**
虚拟购买事件

设置止损止盈策略
 */
package events

import (
	"github.com/locxiang/bitStrategyEngine/dispatcher"
	"time"
	"fmt"
	"strings"
	"os"
	"sync"
	"github.com/locxiang/bitStrategyEngine/util"
)

type VirtualPurchase struct {
	Id        string  //任务唯一id
	SName     string  //策略的名字
	symbol    string  //购买的货币
	Quantity  float64 //购买数量
	buyPrice  float64 //实际购买的价格
	GoodRate  float64 //获利抛售
	BadRate   float64 //止损抛售
	orderID   string  //订单结构体
	sellPrice float64 //实际出售的价格
	sync.Mutex
}

func (p *VirtualPurchase) goodPrice() float64 {
	defer p.Unlock()
	p.Lock()
	return p.buyPrice * (1 + p.GoodRate)
}

func (p *VirtualPurchase) badPrice() float64 {
	defer p.Unlock()
	p.Lock()
	return p.buyPrice * (1 - p.BadRate)
}
func (p *VirtualPurchase) SetSymbol(s string) {
	defer p.Unlock()
	p.Lock()
	p.symbol = s
}

func (p *VirtualPurchase) SetId(s string) {
	defer p.Unlock()
	p.Lock()
	p.Id = s
}

func (p *VirtualPurchase) GetId() string {
	defer p.Unlock()
	p.Lock()
	return p.Id
}

func (p *VirtualPurchase) Handle(strategy dispatcher.Strategy) {
	p.SetSymbol(strategy.GetPool().GetSymbol())

	//购买订单
	p.buy(strategy)

	//循环检测价格，发现达到指标卖出
	go func() {
		tick := time.NewTicker(10 * time.Second)

		for {
			select {
			case <-time.After(500 * time.Millisecond):
				lp := strategy.GetPool().LastPrice()
				if p.checkSell(lp) {
					//出售
					p.sell(strategy)
					//计算收益
					p.CalculatingBenefits()
					tick.Stop()
					return //结束循环
				}
			case <-tick.C:
				lp := strategy.GetPool().LastPrice()
				fmt.Printf("[%s][%s]最新价格：%f  buy:%f , good:%f  bad:%f  %s\n", p.Id, p.orderID, lp, p.buyPrice, p.goodPrice(), p.badPrice(), time.Now())
			}
		}
	}()
}

//判断是否可以卖出
func (p *VirtualPurchase) checkSell(lastPrice float64) bool {

	if p.goodPrice() <= lastPrice {
		return true
	}

	if p.badPrice() >= lastPrice {
		return true
	}

	return false

}

func (p *VirtualPurchase) buy(strategy dispatcher.Strategy) bool {
	defer p.Unlock()
	p.Lock()
	p.buyPrice = strategy.GetPool().LastPrice()
	p.orderID = util.GetRandomString(10)
	fmt.Printf("购买成功：%s  %f\n", p.symbol, p.buyPrice)

	return true
}

func (p *VirtualPurchase) sell(strategy dispatcher.Strategy) bool {
	defer p.Unlock()
	p.Lock()

	p.sellPrice = strategy.GetPool().LastPrice()
	fmt.Printf("出售成功：%s  %f\n", p.symbol, p.sellPrice)
	return true
}

func (p *VirtualPurchase) CalculatingBenefits() float64 {
	defer p.Unlock()
	p.Lock()

	//TODO 千分之2的手续费
	s := (p.sellPrice-p.buyPrice)*p.Quantity - (0.0002 * p.sellPrice) - (0.0002 * p.buyPrice)

	t := time.Now().Format(time.RFC3339)
	str := fmt.Sprintf("[%s]收益：%f,  买：%f, 卖%f ", t, s, p.sellPrice, p.buyPrice)
	fmt.Printf("%s \n", str)
	p.traceFile(str)
	return s
}

//追加写入文件
func (p *VirtualPurchase) traceFile(content string) {
	fd, _ := os.OpenFile("benefits_"+p.Id+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	defer fd.Close()

	fdContent := strings.Join([]string{content, "\n"}, "")
	fd.WriteString(fdContent)

}
