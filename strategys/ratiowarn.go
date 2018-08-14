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
	loop    bool                   // 是否循环
	sleep   time.Duration          //如果命中后 休息策略的时间
	timeOut time.Time              //失效时间
	status  bool                   //策略状态是否开启
	pool    dispatcher.TradePool   //属于的数据池
	events  []dispatcher.EventFace //命中策略后执行的时间
	sync.Mutex
}

func (d *RatioWarn) Name() string {
	return fmt.Sprintf("幅度达到 %f%% 则命中", d.ratio*100, )
}

// 判断此策略是否开启
func (d *RatioWarn) Valid() bool {
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

// 开始休息
func (d *RatioWarn) startSleep() {
	fmt.Printf("开始休息：%s \n", d.sleep)
	go func() {
		time.Sleep(d.sleep)
		d.SetStatus(true)
	}()
}

func (d *RatioWarn) GetPool() dispatcher.TradePool {
	return d.pool
}

//检查策略是否命中
func (d *RatioWarn) Check(p dispatcher.TradePool) bool {
	ratio := p.Ratio()

	lastPrice := p.LastPrice()
	firstPrice := p.FirstPrice()

	fmt.Printf("时间范围:%s, 价格:%f-%f ， 变化率：%f%%   策略变化率：%f%%\n", p.GetDuration(), lastPrice, firstPrice, ratio*100, d.ratio*100)

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
	if d.loop {
		d.startSleep()
	} else {
		fmt.Printf("使命完成，结束策略：  %+v \n", d)
		d.pool.UnregisterStrategy(d)
	}
}

//注册事件
func (d *RatioWarn) RegisterEvent(e dispatcher.EventFace) {
	d.events = append(d.events, e)
}

//循环策略
func NewLoopRatioWarn(ratio float64, sleep time.Duration, timeout time.Time, pool dispatcher.TradePool) {
	// 获取数据池
	e := &RatioWarn{
		ratio:   ratio,
		loop:    true,
		timeOut: timeout,
		status:  true,
		sleep:   sleep,
		pool:    pool,
	}

	pool.RegisterStrategy(e)
	fmt.Printf("创建循环策略：%+v \n", e)
}

//一次性策略
func NewDisposableRatioWarn(ratio float64, timeout time.Time, pool dispatcher.TradePool) (*RatioWarn) {

	// 获取数据池
	e := &RatioWarn{
		ratio:   ratio,
		loop:    false,
		timeOut: timeout,
		status:  true,
		pool:    pool,
	}

	pool.RegisterStrategy(e)
	fmt.Printf("创建策略：%+v \n", e)

	//检查
	return e
}
