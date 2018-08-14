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
)

var cfg *config.Config

func init() {
	var err error
	cfg, err = config.Load("config.ini")
	if err != nil {
		panic(err)
	}
}

func main() {

	//加载数据池
	btcpool := &tradepool.TradePool{
		Symbol:   "BTCUSDT",
		Duration: time.Second * 5,
	}
	dispatcher.RegisterPools(btcpool)

	ethpool := &tradepool.TradePool{
		Symbol:   "ETHUSDT",
		Duration: time.Second * 10,
	}
	dispatcher.RegisterPools(ethpool)

	//加载策略
	strategys.NewDisposableRatioWarn(0.0008, time.Time{}, btcpool)
	timeout := time.Now().Add(30 * time.Second)
	strategys.NewLoopRatioWarn(0.0008, time.Second*5, timeout, btcpool)
	strategys.NewLoopRatioWarn(-0.0008, time.Second*5, timeout, ethpool)

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
