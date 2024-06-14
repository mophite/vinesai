package main

import (
	"vinesai/internel/ava"
	"vinesai/internel/ava/_example/tutorial/internal/ipc"
	"vinesai/internel/ava/_example/tutorial/proto/phello"
	"vinesai/internel/ava/_example/tutorial/proto/pim"

	"go.etcd.io/etcd/client/v3"
	"vinesai/internel/ava/_example/tutorial/app/api/api.hello/hello"
	"vinesai/internel/ava/_example/tutorial/app/api/api.hello/hijack"
	"vinesai/internel/ava/_example/tutorial/app/api/api.hello/http"
	"vinesai/internel/ava/_example/tutorial/app/api/api.hello/im"
)

// ```shell
// curl -h "Content-Type:application/json" -X POST -d '{"ping": "ping"}' http://127.0.0.1:10000/hello/say/hi
// ```
func main() {
	ava.SetupService(
		ava.HttpGetRootPathRedirect("/hello/say/hihttp"),
		ava.Namespace("api.hello"),
		ava.HttpApiAdd("0.0.0.0:10000"),
		ava.TCPApiPort(10001),
		ava.WssApiAddr("0.0.0.0:10002", "/hello"),
		ava.Hijacker(hijack.HijackWriter),
		ava.EtcdConfig(&clientv3.Config{Endpoints: []string{"127.0.0.1:2379"}}),
	)

	phello.RegisterSaySrvServer(&hello.Say{})

	// for ava/_example/javascript service
	pim.RegisterImServer(im.NewIm())

	phello.RegisterHttpServer(&http.Http{})

	ipc.InitIpc()

	ava.Run()
}
