package device

import (
	ctx "context"
	"fmt"
	"io"
	"os"
	"runtime"
	"testing"
	"time"

	"vinesai/internel/ava"
	"vinesai/internel/ava/_example/tutorial/proto/phello"

	"github.com/jjeffcaii/reactor-go/scheduler"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx"
	"github.com/rsocket/rsocket-go/rx/flux"
)

// Other language encapsulation request api reference
func init() {
	//newClient()
}

// srv.hello/hello/hellosrv/saychannelc5kfvl6g10l48q7pjss0{"Content-Type":"application/json","X-Api-Version":"v1.0.0"}
func TestRequestChannel(t *testing.T) {

	meta, _ := ava.EncodeMetadata(
		"srv.hello",
		"/hello/saysrv/channel",
		"c5kep5mg10l34dfgudkg",
		map[string]string{"X-Api-Version": "v1.0.0", "Content-Type": "application/json"},
	)

	var (
		sendPayload = make(chan payload.Payload, 10)
	)

	var req = make(chan *phello.SayReq)

	go func() {
		for i := 0; i < 3; i++ {
			req <- &phello.SayReq{
				Ping: "ping",
			}
		}

		//must be closed
		time.Sleep(time.Second * 2)
		//close will close socket
		close(req)
	}()

	go func() {
		sendPayload <- payload.New(meta.Payload(), nil)

	QUIT:
		for {
			select {
			case d, ok := <-req:
				if ok {
					sendPayload <- payload.New(ava.MustMarshal(d), nil)
				} else {
					close(sendPayload)
					break QUIT
				}
			}
		}

	}()

	gClient := newClient()

	var (
		f = gClient.RequestChannel(
			flux.Create(
				func(ctx ctx.Context, s flux.Sink) {
					go func() {
					loop:
						for {
							select {
							case <-ctx.Done():
								s.Error(ctx.Err())
								break loop
							case p, ok := <-sendPayload:
								if ok {
									s.Next(p)
								} else {
									s.Complete()
									break loop
								}
							}
						}
					}()
				},
			),
		)
	)

	var done = make(chan struct{})
	f.
		SubscribeOn(scheduler.Parallel()).
		DoFinally(
			func(s rx.SignalType) {
				//todo handler rx.SignalType
				ava.Debug("DoFinally")
				done <- struct{}{}
			},
		).
		Subscribe(
			ctx.Background(),
			rx.OnNext(
				func(p payload.Payload) error {
					ava.Infof("from server |data=%s", p.DataUTF8())
					return nil
				},
			),
			rx.OnError(
				func(err error) {
					ava.Error(err)
				},
			),
		)
	<-done
}

func newClient() rsocket.Client {
	client, err := rsocket.
		Connect().
		Scheduler(
			scheduler.NewElastic(runtime.NumCPU()<<8),
			scheduler.NewElastic(runtime.NumCPU()<<8),
		). //set scheduler to best
		KeepAlive(time.Second*5, time.Second*5, 1).
		ConnectTimeout(time.Second * 5).
		OnConnect(
			func(client rsocket.Client, err error) { //handler when connect success
				ava.Debug("connected success")
			},
		).
		OnClose(
			func(err error) { //when net occur some error,it's will be callback the error server ip address
				ava.Error(err)
			},
		).
		Transport(rsocket.TCPClient().SetAddr("172.23.26.2:20001").Build()). //setup transport and start
		//Transport(rsocket.WebsocketClient().SetURL("ws://0.0.0.0:7777/test/wss").Build()). //setup transport and start
		Start(ctx.TODO())
	if err != nil {
		panic(err)
	}
	return client
}

func TestAsr(t *testing.T) {
	f, err := os.Open("./test.wav")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	fd, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	rsp, err := asr(fd)
	rsp.ToJsonString()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(*rsp.Response.Result)
}
