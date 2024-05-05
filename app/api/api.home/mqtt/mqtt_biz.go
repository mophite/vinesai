package mqtt

import (
	"errors"
	"fmt"
	"time"
	"vinesai/internel/ava"
	"vinesai/internel/db"
	"vinesai/internel/db/db_hub"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"gorm.io/gorm"
)

var deviceReportTopic = "home/report/device"
var deviceControlTopic = "home/control/device/%s/%s" //独立设备id的topic

var client mqtt.Client

func Chaos() error {

	opts := mqtt.NewClientOptions()

	opts.AddBroker("tcp://127.0.0.1:1883")
	opts.SetClientID(ava.RandString(12))
	opts.SetUsername("root")
	opts.SetPassword("000000")
	opts.SetKeepAlive(600 * time.Second)
	opts.SetPingTimeout(5 * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetDefaultPublishHandler(f)
	opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		ava.Error(err)
	})

	// 启动一个链接
	client = mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	time.Sleep(time.Second)
	go mqttReportSubscribe()

	return nil
}

var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("主题: %s\n", msg.Topic())
	fmt.Printf("信息: %s\n", string(msg.Payload()))
}

// 发布消息给指定的客户端
func mqttPublish(c *ava.Context, deviceId, userId, data string) {

	var topic = fmt.Sprintf(deviceControlTopic, userId, deviceId)
	c.Debugf("mqttPublish |topic=%s |data=%s", topic, data)
	//retained false服务器不保留消息
	token := client.Publish(topic, 0, false, data)
	//保证消息推送成功
	go func() {
		if !(token.Wait() && token.Error() == nil) {
			ava.Debugf("publish failure |deviceId=%s |err=%v", deviceId, token.Error())
			ava.Error(token.Error())
		}
	}()
}

// 测试版本智能插座
// 订阅客户端发送的消息:数据上报
// 将数据状态同步到数据库中
func mqttReportSubscribe() {
	token := client.Subscribe(deviceReportTopic, byte(0), func(c mqtt.Client, message mqtt.Message) {
		ava.Debugf("subscribe topic=%s|payload=%s", deviceReportTopic, string(message.Payload()))

		//todo 不同品牌的设备不同的topic
		var adaptor db_hub.SocketMiniV2
		err := ava.Unmarshal(message.Payload(), &adaptor)
		if err != nil {
			ava.Error(err)
			return
		}

		device := adaptor.Adaptor2Device()

		err = db.GMysql.Transaction(func(tx *gorm.DB) error {

			var d db_hub.Device
			err = tx.Table(db_hub.TableDeviceList).
				Where("device_id=? AND user_id=?", device.DeviceID, "123").
				Take(&d).
				Error

			if errors.Is(err, gorm.ErrRecordNotFound) {
				//数据不存在,插入数据
				return tx.Table(db_hub.TableDeviceList).Create(&d).Error
			}

			//存在则更新数据
			updates := make(map[string]interface{}, 10)
			if device.DeviceZn != "" {
				updates["device_zn"] = device.DeviceZn
			}
			if device.DeviceEn != "" {
				updates["device_en"] = device.DeviceEn
			}
			if device.DeviceDes != "" {
				updates["device_des"] = device.DeviceDes
			}
			if device.Control != 0 {
				updates["control"] = device.Control
			}
			return db.GMysql.Table(db_hub.TableDeviceList).
				Where("device_id=? AND user_id=?", device.DeviceID, "123").
				Updates(updates).Error

		})

		if err != nil {
			ava.Error(err)
		}
	})

	if token.Wait() && token.Error() != nil {
		panic(fmt.Sprintf("sbscribe |topic=%s |err=%v", deviceReportTopic, token.Error()))
	}

}

var botTmp = `你现在是一个智能家居中控系统。根据我提供的设备数据,通过英文字段和值来分析设备的当前状态,并根据我描述的场景来控制它们。
场景控制指令下达后,请严格按照以下JSON格式,在唯一的代码块中输出结果,不要有其他多余内容:
{
"result": [
{
"user_id": "123",
"device_type": 1,
"device_id": "8CCE4E522308",
"control": 1
}
],
"message": "您好,主人!卧室灯已经关闭。今晚祝您做个好梦~"
}
字段说明如下,未提及的字段直接忽略:
device_type:设备类型
device_zn:设备中文名称
device_en:设备英文名称
device_id:设备ID
user_id:所属用户ID
control:开关状态,1为断电,2为通电
注意事项:

1、user_id、device_id、device_type 是必须字段,不可修改它们的值,其他字段只保留有变更的。
2、将处理后的设备数据放入 result 数组中。
3、以幽默活泼的口吻回应场景执行结果,在 message 字段中说明调整的设备。

我会提供设备数据如下:
%s
请分析以上数据,等待我的控制指令。一旦收到指令,即刻按要求输出JSON结果。`
