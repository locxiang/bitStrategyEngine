package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/binance-exchange/go-binance"
	"github.com/locxiang/bitStrategyEngine/config"
	"time"
	"github.com/locxiang/bitStrategyEngine/strategys"
	"github.com/locxiang/bitStrategyEngine/tradepool"
	"github.com/locxiang/bitStrategyEngine/dispatcher"
	"github.com/locxiang/bitStrategyEngine/events"
	"runtime"
	"strings"
	"github.com/locxiang/bitStrategyEngine/util"
)

var cfg *config.Config

func init() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	var err error
	cfg, err = config.Load("config.ini")
	if err != nil {
		panic(err)
	}
}

func main() {

	//{
	//	sg := strategyGroup{
	//		util.GetRandomString(8), "BTCUSDT", 5 * time.Second, 0.0004, nil, 0.1, 0.004, 0.003,
	//	}
	//	NewStrategyGroup(sg)
	//}
	//
	//{
	//	sg := strategyGroup{
	//		util.GetRandomString(8), "BTCUSDT", 5 * time.Second, 0.0008, nil, 0.1, 0.004, 0.003,
	//	}
	//	NewStrategyGroup(sg)
	//}

	//{
	//	sg := strategyGroup{
	//		util.GetRandomString(8), "BTCUSDT", 3 * time.Second, 0.0005, nil, 0.1, 0.003, 0.002,
	//	}
	//	NewStrategyGroup(sg)
	//}

	{
		sg := strategyGroup{
			util.GetRandomString(8), "BTCUSDT", 10 * time.Second, 0.0001, nil, 0.1, 0.001, 0.001,
		}
		NewStrategyGroup(sg)
	}

	ctx, cancelCtx := context.WithCancel(context.Background())
	b := NewBinanceService(ctx)
	fmt.Printf("连接交易所成功 \n")

	//生成假订单
	//
	//err := b.NewOrderTest(binance.NewOrderRequest{
	//	Symbol:      "BTCUSDT",
	//	Quantity:    1,
	//	Price:       999,
	//	Side:        binance.SideBuy,
	//	TimeInForce: binance.GTC,
	//	Type:        binance.TypeMarket,
	//	Timestamp:   time.Now(),
	//})
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(err)

	fmt.Printf("创建websocket连接... ")
	kech, done, err := b.TradeWebsocket(binance.TradeWebsocketRequest{
		Symbol: "BTCUSDT",
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf(" & 连接成功 \n")
	go func() {
	WS_OVER:
		for {
			select {
			case ke := <-kech:
				dispatcher.Disp.EventAggTrade <- ke
			case <-done:
				fmt.Printf("收到关闭命令 \n")
				break WS_OVER
			}
		}
	}()

	<-done
	fmt.Println("exit")
	cancelCtx()
	fmt.Println("waiting for signal")
	return

	ds, err := b.StartUserDataStream()
	if err != nil {
		panic(err)
	}
	fmt.Printf("StartUserDataStream %+v\n", ds)

	err = b.KeepAliveUserDataStream(ds)
	if err != nil {
		panic(err)
	}

	err = b.CloseUserDataStream(ds)
	if err != nil {
		panic(err)
	}
}

func NewBinanceService(ctx context.Context) binance.Service {
	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = level.NewFilter(logger, level.AllowAll())
	logger = log.With(logger, "time", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	hmacSigner := &binance.HmacSigner{
		Key: []byte(cfg.SecretKey),
	}

	// use second return value for cancelling request
	binanceService := binance.NewAPIService(
		cfg.Url,
		cfg.ApiKey,
		hmacSigner,
		logger,
		ctx,
	)
	b := binance.NewBinance(binanceService)

	return b
}

type strategyGroup struct {
	Id                          string
	symbol                      string
	poolTime                    time.Duration
	ratio                       float64
	timeout                     *time.Time
	quantity, goodRate, badRate float64
}

func (sg *strategyGroup) String() string {
	str := fmt.Sprintf("[%s]%s:%s %f,%s  quantity:%f,goodRate:%f,badRate:%f", sg.Id, sg.symbol, sg.poolTime, sg.ratio, sg.timeout, sg.quantity, sg.goodRate, sg.badRate)
	return str
}

func NewStrategyGroup(sg strategyGroup) {

	fmt.Printf("开启策略，名字：%s \n", sg.String())
	//加载数据池
	btcpool := &tradepool.TradePool{
		Symbol:   sg.symbol,
		Duration: sg.poolTime,
	}
	dispatcher.RegisterPools(btcpool)

	//加载策略
	//timeout := time.Now().Add(30 * time.Second)
	strategyBTC, done := strategys.NewRatioWarn(sg.ratio, sg.timeout, btcpool)
	go func() {
		select {
		case <-done:
			fmt.Printf("done %s\n", sg.String())
			time.Sleep(3 * time.Second)

			sg.Id = util.GetRandomString(8)
			NewStrategyGroup(sg)
		}
	}()

	//加载事件
	eventVirtualPurchase := &events.VirtualPurchase{
		Id:       sg.Id,
		SName:    sg.String(),
		Quantity: sg.quantity,
		GoodRate: sg.goodRate,
		BadRate:  sg.badRate,
	}
	strategyBTC.RegisterEvent(eventVirtualPurchase)

	//TODO 保存组合id
	fd, _ := os.OpenFile("strategy_group.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	defer fd.Close()

	content := fmt.Sprintf("ID:%s strategy:%s", sg.Id, sg.String())
	fdContent := strings.Join([]string{content, "\n"}, "")
	fd.WriteString(fdContent)
}
