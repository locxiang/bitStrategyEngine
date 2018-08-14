package config

import (
	"github.com/go-ini/ini"
)

type Exchange struct {
	Type      string `ini:"Type"`
	Url       string
	ApiKey    string
	SecretKey string
	Spot      []string
}

type Config struct {
	Exchange `ini:"exchange"`
}

func Load(filename string) (*Config, error) {
	cfg, err := ini.Load(filename)
	if err != nil {
		return nil, err
	}

	var env = &Config{}
	err = cfg.MapTo(env)
	
	return env, nil
}
