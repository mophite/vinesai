package main

import (
	"github.com/coreos/etcd/clientv3"
	"vinesai/app/api/srv.hub/gpt"
	"vinesai/internel/ava"
	"vinesai/internel/config"
	"vinesai/proto/phub"
)

func main() {

	ava.SetupService(
		ava.Namespace("srv.hub"),
		ava.EtcdConfig(&clientv3.Config{Endpoints: []string{"127.0.0.1:2379"}}),
		ava.ConfigOption(ava.Chaos(config.Chaos, gpt.ChaosOpenAI)),
	)

	phub.RegisterChatServer(&gpt.Gpt{})

	ava.Run()
}
