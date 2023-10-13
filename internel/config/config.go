package config

import (
	"vinesai/internel/ava"
	"vinesai/internel/db"
)

var GConfig *config

type config struct {
	Redis  redis  `json:"redis"`
	Mysql  mysql  `json:"mysql"`
	OpenAI openAI `json:"openai"`
}

//go:generate etcdctl put  configava/v1.0.0/public/ava.redis "{ \"address\":\"127.0.0.1:6379\", \"password\":\"\" }"
type redis struct {
	Address  string `json:"address"`
	Password string `json:"password"`
}

//go:generate etcdctl put  configava/v1.0.0/public/ava.mysql "{ \"dsn\":\"root:12345678@tcp(127.0.0.1:3306)/vinesai?charset=utf8mb4&loc=Local\"}"
type mysql struct {
	Dsn string
}

//go:generate etcdctl put  configava/v1.0.0/private/openai "{ \"base_url\":\"https://api.openai-proxy.com/v1/\",\"key\":\"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\",\"temperature\":0.1,\"top_p\":0}"
type openAI struct {
	BaseURL     string  `json:"base_url"`
	Key         string  `json:"key"`
	Temperature float32 `json:"temperature"`
	TopP        float32 `json:"top_p"`
}

func ChaosOpenAI() error {
	if GConfig == nil {
		GConfig = new(config)
	}

	var o openAI
	err := ava.ConfigDecPrivate("openai", &o)
	if err != nil {
		ava.Error(err)
		return err
	}

	GConfig.OpenAI = o

	return nil
}

func ChaosDB() error {
	if GConfig == nil {
		GConfig = new(config)
	}

	var r redis
	err := ava.ConfigDecPublic("redis", &r)
	if err != nil {
		ava.Error(err)
		return err
	}

	ava.Debugf("redis |data=%v", r)

	GConfig.Redis = r

	//初始化redis
	err = db.ChaosRedis(r.Address)
	if err != nil {
		ava.Error(err)
		return err
	}

	var m mysql
	err = ava.ConfigDecPublic("mysql", &m)
	if err != nil {
		ava.Error(err)
		return err
	}

	GConfig.Mysql = m

	//初始化mysql
	err = db.ChaosMysql(m.Dsn)
	if err != nil {
		ava.Error(err)
		return err
	}

	//初始化openai
	var o openAI
	err = ava.ConfigDecPrivate("openai", &o)
	if err != nil {
		ava.Error(err)
		return err
	}

	GConfig.OpenAI = o

	return nil
}
