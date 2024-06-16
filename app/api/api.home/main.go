package main

import (
	"net/http"
	"time"
	"vinesai/internel/ava"
	"vinesai/internel/config"
	"vinesai/proto/pha"
	"vinesai/proto/pmini"

	"go.etcd.io/etcd/client/v3"
	"vinesai/app/api/api.home/homeassistant"
	"vinesai/app/api/api.home/miniprogram"
	"vinesai/app/api/api.home/mqtt"
	"vinesai/app/api/api.home/user"
)

func main() {

	ava.SetupService(
		ava.Namespace("api.home"),
		ava.HttpApiAdd("0.0.0.0:10005"),
		//ava.TCPApiPort(10001),
		//ava.WssApiAddr("0.0.0.0:10002", "/ws"),
		ava.EtcdConfig(&clientv3.Config{Endpoints: []string{"127.0.0.1:2379"}}),
		//ava.WatchDog(oauth2.Oauth),
		ava.ConfigOption(
			ava.Chaos(
				config.ChaosDB,
				config.ChaosOpenAI,
				miniprogram.ChaosOpenAI,
				mqtt.Chaos,
			)),
		//ava.Cors(lib.Cors()),
	)

	pmini.RegisterChat4MiniServer(miniprogram.NewMini())
	pmini.RegisterDeviceControlServer(&mqtt.MqttHub{})
	pha.RegisterLlmServer(&homeassistant.HomeAssistant{})
	pha.RegisterUserServer(&user.User{RemoteIp: "43.139.244.233"})

	go func() {
		//启动原生的websocket
		http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
			miniprogram.ServeWs(w, r)
		})
		server := &http.Server{
			Addr:              ":10002",
			ReadHeaderTimeout: 5 * time.Second,
		}
		err := server.ListenAndServe()
		if err != nil {
			ava.Error(err)
		}
	}()

	ava.Run()
}
