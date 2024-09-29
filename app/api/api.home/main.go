package main

import (
	"vinesai/internel/ava"
	"vinesai/internel/config"
	"vinesai/proto/ptuya"

	"go.etcd.io/etcd/client/v3"
	"vinesai/app/api/api.home/tuya"
)

func main() {

	ava.SetupService(
		ava.Namespace("api.home"),
		ava.HttpApiAdd("0.0.0.0:10010"),
		//ava.TCPApiPort(10001),
		//ava.WssApiAddr("0.0.0.0:10002", "/ws"),
		ava.EtcdConfig(&clientv3.Config{Endpoints: []string{"127.0.0.1:2379"}}),
		//ava.WatchDog(tuya.Authorization),
		ava.ConfigOption(
			ava.Chaos(
				config.ChaosRedis,
			)),
		//ava.Cors(lib.Cors()),
	)

	ptuya.RegisterTuyaServer(&tuya.Tuya{})

	ava.Run()
}
