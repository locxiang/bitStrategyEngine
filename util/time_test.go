package util

import (
	"testing"
	time "time"
	"fmt"
)

func TestTimeDiff(t *testing.T) {

	//Add方法和Sub方法是相反的，获取t0和t1的时间距离d是使用Sub，将t0加d获取t1就是使用Add方法
	k := time.Now()

	//一天之前
	d, _ := time.ParseDuration("-24h")

	k2 := k.Add(d)

	 f := k.Sub(k2)
	fmt.Printf("%s", f)

}