package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"vinesai/internel/ava"
	"vinesai/internel/ipc"
	"vinesai/proto/phub"

	"go.etcd.io/etcd/client/v3"
	"vinesai/app/api/api.hub/device"
	"vinesai/app/api/api.hub/oauth2"
)

func main() {
	// Setup the service
	ava.SetupService(
		ava.Namespace("api.hub"),
		ava.HttpApiAdd("0.0.0.0:10000"),
		//ava.TCPApiPort(10001),
		//ava.WssApiAddr("0.0.0.0:10002", "/hub"),
		ava.EtcdConfig(&clientv3.Config{Endpoints: []string{"127.0.0.1:2379"}}),
		ava.WatchDog(oauth2.Oauth),
		//ava.ConfigOption(ava.Chaos(config.ChaosDB)),
	)

	phub.RegisterDeviceServer(&device.DevicesHub{})
	phub.RegisterOauthServer(&oauth2.Oauth2{})

	ipc.InitIpc()

	go func() {
		// Run the service
		ava.Run()
	}()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		shutdown()
	}
}

func shutdown() {
	fmt.Println("-------------------exit------------------")
	time.Sleep(time.Second * 2)
}
