package main

import (
	"vinesai/internel/ava"
	"vinesai/internel/config"
	"vinesai/proto/phub"

	"github.com/coreos/etcd/clientv3"
	"vinesai/app/api/srv.hub/gpt"
)

func main() {

	ava.SetupService(
		ava.TCPApiPort(30001),
		ava.Namespace("srv.hub"),
		ava.EtcdConfig(&clientv3.Config{Endpoints: []string{"175.178.164.183:2379"}}),
		ava.ConfigOption(ava.Chaos(config.ChaosOpenAI, gpt.ChaosOpenAI)),
	)

	phub.RegisterChatServer(&gpt.Gpt{})

	ava.Run()
}
