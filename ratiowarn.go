package main

import (
	"time"
	"fmt"
)

//跌幅比预警
type RatioWarn struct {
	symbol   string
	duration time.Duration //在一定时间内
	ratio    float64       // 涨/跌幅百分比
	loop     bool          // 是否循环
	sleep    time.Duration //如果命中后 休息策略的时间
	timeOut  *time.Time    //失效时间
	status   bool          //策略状态是否开启
}

// 判断次策略是否有效
func (d *RatioWarn) Valid() bool {
	// 过期
	if d.TimeOut().Before(time.Now()) {
		//过期了,就跳过
		fmt.Printf("此策略过期了:%+v \n", d)
		return false
	}

	return d.status
}

// 开始休息
func (d *RatioWarn) startSleep() {
	fmt.Printf("开始休息：%s \n", d.sleep)
	d.status = false
	go func() {
		time.Sleep(d.sleep)
		d.status = true
	}()
}

//检查策略是否命中
func (d *RatioWarn) Check() bool {
	ratio := d.Pool().Ratio()

	e := d.Pool()

	lastPrice, _ := e.lastPrice()
	firstPrice, _ := e.firstPrice()

	fmt.Printf("时间范围:%s, 价格:%f-%f ， 变化率：%f%%   策略变化率：%f%%\n", e.duration, lastPrice, firstPrice, ratio*100, d.ratio*100)

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

func (d *RatioWarn) Handle() {
	fmt.Printf("￥￥￥￥￥￥￥￥￥￥￥￥￥￥￥￥￥命中策略：%+v \n", d)

	// 休息策略
	d.status = false

	if d.loop {
		if d.sleep > 0 {
			d.startSleep()
		} else {
			d.status = true
		}
	}
}

func (d *RatioWarn) TimeOut() *time.Time {
	//如果为空 就创造一个不过期的时间
	if d.timeOut == nil {
		s := time.Now().Add(10 * time.Second)
		return &s
	}

	return d.timeOut
}

//获取数据池
func (d *RatioWarn) Pool() *tradePool {
	pool := GetPool(d.symbol, d.duration)
	return pool
}

//循环策略
func NewLoopRatioWarn(symbol string, duration time.Duration, ratio float64, sleep time.Duration, timeout *time.Time) {
	// 获取数据池
	e := &RatioWarn{
		symbol:   symbol,
		duration: duration,
		ratio:    ratio,
		loop:     true,
		timeOut:  timeout,
		status:   true,
		sleep:    sleep,
	}

	pool := e.Pool()
	pool.AddStrategy(e)

	fmt.Printf("创建循环策略：%+v \n", e)
}

//一次性策略
func NewDisposableRatioWarn(symbol string, duration time.Duration, ratio float64, timeout *time.Time) (*RatioWarn) {
	// 获取数据池
	e := &RatioWarn{
		symbol:   symbol,
		duration: duration,
		ratio:    ratio,
		loop:     false,
		timeOut:  timeout,
		status:   true,
	}

	pool := e.Pool()
	pool.AddStrategy(e)

	fmt.Printf("创建策略：%+v \n", e)

	//检查
	return e
}
