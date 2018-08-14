package config

import "testing"

func TestLoad(t *testing.T) {
	cfg, err := Load("/Users/xh/golib/src/github.com/locxiang/btcStrategyEngine/config.ini")
	t.Log(cfg, err)

}
