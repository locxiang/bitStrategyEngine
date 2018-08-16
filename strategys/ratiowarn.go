package strategys

import (
	"time"
	"fmt"
	"github.com/locxiang/bitStrategyEngine/dispatcher"
	"sync"
)

/**
跌幅比预警
次策略的逻辑是
在时间n范围内 波动比率达到x， 则进行策略

 */
type RatioWarn struct {
	ratio   float64                // 涨/跌幅百分比
	timeOut *time.Time             //失效时间
	status  bool                   //策略状态是否开启
	pool    dispatcher.TradePool   //属于的数据池
	events  []dispatcher.EventFace //命中策略后执行的时间
	done    chan struct{}          //结束事件
	sync.Mutex
}

func (d *RatioWarn) Name() string {
	defer d.Unlock()
	d.Lock()
	return fmt.Sprintf("幅度达到 %f%% 则命中", d.ratio*100, )
}

// 判断此策略是否开启
func (d *RatioWarn) Valid() bool {

	if d.timeOut == nil {
		return d.GetStatus()
	}
	// 过期
	if d.timeOut.Before(time.Now()) {
		//过期了,就跳过
		fmt.Printf("此策略过期了:%+v \n", d)
		d.GetPool().UnregisterStrategy(d)
		return false
	}

	return d.GetStatus()
}

func (d *RatioWarn) GetStatus() bool {
	defer d.Unlock()
	d.Lock()
	return d.status
}

func (d *RatioWarn) SetStatus(b bool) {
	defer d.Unlock()
	d.Lock()
	d.status = b
}

func (d *RatioWarn) GetPool() dispatcher.TradePool {
	return d.pool
}

//检查策略是否命中
func (d *RatioWarn) Check() bool {

	ratio := d.GetPool().Ratio()
	//
	//lastPrice := d.GetPool().LastPrice()
	//firstPrice := d.GetPool().FirstPrice()
	//fmt.Printf("时间范围:%s, 价格:%f-%f ， 变化率：%f%%   策略变化率：%f%%\n", d.GetPool().GetDuration(), lastPrice, firstPrice, ratio*100, d.ratio*100)

	// 涨幅命中
	if ratio > 0 && ratio >= d.ratio {
		return true
	}

	//跌幅命中
	if ratio < 0 && ratio <= d.ratio {
		return true
	}

	return false
}

//命中策略执行事件
func (d *RatioWarn) HandleEvents() {
	fmt.Printf("￥￥￥￥￥￥￥￥￥￥￥￥￥￥￥￥￥命中策略：%s \n", d.Name())
	d.doneCheck()

	// 循环执行策略
	for _, e := range d.events {
		e.Handle(d)
	}

}

//命中策略
func (d *RatioWarn) doneCheck() {
	// 休息策略
	d.SetStatus(false)
	d.done <- struct{}{}
}

//注册事件
func (d *RatioWarn) RegisterEvent(e dispatcher.EventFace) {
	defer d.Unlock()
	d.Lock()

	d.events = append(d.events, e)
}

//一次性策略
func NewRatioWarn(ratio float64, timeout *time.Time, pool dispatcher.TradePool) (dispatcher.Strategy, <-chan struct{}) {

	done := make(chan struct{})
	// 获取数据池
	e := &RatioWarn{
		ratio:   ratio,
		timeOut: timeout,
		status:  true,
		pool:    pool,
		done:    done,
	}

	pool.RegisterStrategy(e)
	//fmt.Printf("创建策略：%+v \n", e)

	//检查
	return e, done
}
