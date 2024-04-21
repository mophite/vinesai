package main

import (
	"vinesai/internel/ava"
	"vinesai/internel/config"
	"vinesai/proto/phub"

	"github.com/coreos/etcd/clientv3"
	"vinesai/app/srv/srv.hub/chatgpt4esp"
)

func main() {

	ava.SetupService(
		//ava.EndpointIp("43.132.184.162"),
		//ava.TCPApiPort(30001),
		ava.Namespace("srv.hub"),
		ava.EtcdConfig(&clientv3.Config{Endpoints: []string{"127.0.0.1:2379"}}),
		ava.ConfigOption(
			ava.Chaos(
				config.ChaosOpenAI,
				chatgpt4esp.ChaosOpenAI,
			)),
	)

	phub.RegisterChatServer(&chatgpt4esp.Gpt{})

	ava.Run()
}
