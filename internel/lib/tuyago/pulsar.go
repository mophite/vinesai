package tuyago

import (
	"context"
	"encoding/base64"
	"fmt"
	"vinesai/internel/ava"
	"vinesai/internel/x"

	"github.com/apache/pulsar-client-go/pulsar"
)

var pulsarClient pulsar.Client

var pulsarConsumer pulsar.Consumer

var pulsarProducer pulsar.Producer

func Chaos() error {
	var l, err = ava.LocalIp()
	if err != nil {
		panic(err)
	}

	pulsarClient, err = pulsar.NewClient(pulsar.ClientOptions{
		URL:            defaultMsgHost,
		Authentication: newAuthProvider(defaultClientID, defaultKey),
		ListenerName:   l,
	})

	if err != nil {
		panic(err)
	}

	pulsarConsumer, err = pulsarClient.Subscribe(pulsar.ConsumerOptions{
		Topic:            defaultTopic,
		SubscriptionName: defaultSubscriptionName,
		Type:             pulsar.Failover,
	})

	if err != nil {
		panic(err)
	}

	//pulsarProducer, err = pulsarClient.CreateProducer(pulsar.ProducerOptions{
	//	Topic: defaultTopic,
	//})
	//
	//if err != nil {
	//	panic(err)
	//}

	//启动10个协程处理消费者数据
	for i := 0; i < 10; i++ {
		go func() {
			for {
				var c = ava.Background()

				fmt.Println("----1-")
				msg, err := pulsarConsumer.Receive(context.Background())
				if err != nil {
					c.Error(err)
					continue
				}

				//通知消息队列这条消息已经处理
				err = pulsarConsumer.Ack(msg)
				if err != nil {
					c.Error(err)
				}

				//解密数据
				var data protocolData
				err = x.MustUnmarshal(msg.Payload(), &data)
				if err != nil {
					c.Error(err)
					continue
				}

				de, err := base64.StdEncoding.DecodeString(data.Data)
				if err != nil {
					c.Error(err)
					continue
				}

				payload := x.EcbDecrypt(de, []byte(defaultKey[8:24]))
				c.Debugf("Pulsar |EcbDecrypt data=%s", string(payload))

				//获取bizCode
				bizCodeIn := x.Json.Get(payload, "bizCode").ToString()
				if bizCodeIn == "" {
					continue
				}

				//通过协议号找到对应的处理函数
				invoke(c, bizCodeIn, payload)
			}
		}()
	}

	return nil
}

func Producer(c *ava.Context, p *protocolData) {

	var message = &pulsar.ProducerMessage{
		Payload: x.MustMarshal(p),
	}
	msgID, err := pulsarProducer.Send(context.Background(), message)
	if err != nil {
		c.Error(err)
		return
	}

	fmt.Println("---", msgID)
}

type protocolData struct {
	Protocol int    `json:"protocol"` //协议号
	Pv       string `json:"pv"`       //协议版本
	T        int    `json:"t"`        //时间戳
	Data     string `json:"data"`     //业务数据加密
	Sign     string `json:"sign"`     //签名校验,接收不需要，发送的时候需要
}

func ClosePulsarClient() {
	pulsarClient.Close()
	pulsarConsumer.Close()
}
