package main

import (
	"github.com/coreos/etcd/clientv3"
	"vinesai/internel/ava"
	"vinesai/internel/ava/_example/tutorial/app/srv/srv.hello/hello"
	"vinesai/internel/ava/_example/tutorial/proto/phello"
)

func main() {
	ava.SetupService(
		ava.HttpApiAdd("0.0.0.0:10000"),
		ava.Namespace("srv.hello"),
		ava.TCPApiPort(20001),
		ava.EtcdConfig(&clientv3.Config{Endpoints: []string{"127.0.0.1:2379"}}),
	)

	phello.RegisterSaySrvServer(&hello.Say{})

	ava.Run()
}
