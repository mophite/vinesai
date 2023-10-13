package main

import (
	"vinesai/internel/ava"
	"vinesai/internel/config"
	"vinesai/internel/ipc"
	"vinesai/proto/phub"

	"github.com/coreos/etcd/clientv3"
	"vinesai/app/api/api.hub/device"
	"vinesai/app/api/api.hub/oauth2"
)

func main() {

	ava.SetupService(
		ava.Namespace("api.hub"),
		ava.HttpApiAdd("0.0.0.0:10000"),
		ava.TCPApiPort(10001),
		ava.WssApiAddr("0.0.0.0:10002", "/ws"),
		ava.EtcdConfig(&clientv3.Config{Endpoints: []string{"127.0.0.1:2379"}}),
		ava.WatchDog(oauth2.Oauth),
		ava.ConfigOption(ava.Chaos(config.ChaosDB)),
	)

	phub.RegisterDeviceServer(&device.DevicesHub{})
	phub.RegisterOauthServer(&oauth2.Oauth2{})

	ipc.InitIpc()

	ava.Run()
}
