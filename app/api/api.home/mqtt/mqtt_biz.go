package mqtt

import (
	"fmt"
	"time"
	"vinesai/internel/ava"
	"vinesai/internel/db"
	"vinesai/internel/db/db_hub"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var deviceReportTopic = "home/report/device"
var deviceControlTopic = "home/control/device/%s/%s" //独立设备id的topic

var client mqtt.Client

func Chaos() error {

	opts := mqtt.NewClientOptions()

	opts.AddBroker("tcp://127.0.0.1:1883")
	opts.SetClientID("123")
	opts.SetUsername("root")
	opts.SetPassword("000000")
	opts.SetCleanSession(true)
	opts.SetKeepAlive(600 * time.Second)
	opts.SetPingTimeout(5 * time.Second)
	opts.SetDefaultPublishHandler(f)

	// 启动一个链接
	client = mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	go mqttReportSubscribe()

	return nil
}

var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("主题: %s\n", msg.Topic())
	fmt.Printf("信息: %s\n", msg.Payload())
}

// 发布消息给指定的客户端
func mqttPublish(deviceId, userId, data string) {
	//retained false服务器不保留消息
	token := client.Publish(fmt.Sprintf(deviceControlTopic, userId, deviceId), 0, false, data)
	//保证消息推送成功
	if !(token.Wait() && token.Error() == nil) {
		ava.Debugf("publish failure |deviceId=%s |err=%v", deviceId, token.Error())
		ava.Error(token.Error())
	}
}

// 订阅客户端发送的消息:数据上报
// 将数据状态同步到数据库中
func mqttReportSubscribe() {
	token := client.Subscribe(deviceReportTopic, byte(0), func(c mqtt.Client, message mqtt.Message) {
		ava.Debugf("subscribe |payload=%s", string(message.Payload()))
		//将数据存到mysql中
		var device db_hub.Device
		err := ava.Unmarshal(message.Payload(), &device)
		if err != nil {
			ava.Error(err)
			return
		}

		//消息入库,如果数据库存在该数据直接替换,save方法是强制全部替换整条数据
		err = db.GMysql.Table(db_hub.TableDeviceList).FirstOrCreate(&device, &db_hub.Device{DeviceId: device.DeviceId}).Error
		if err != nil {
			ava.Error(err)
			return
		}
	})

	token.Wait()

	if token.Error() != nil {
		panic(fmt.Sprintf("sbscribe |topic=%s |err=%v", deviceReportTopic, token.Error()))
	}

}

var botTmp = ` 希望你充当一个智能家居的中控系统，在//*数组*//中是智能家居里面的设备清单，你可以根据我提出的
智能家居场景控制他们。当我向你说出场景时，你要按照下面的数据规则格式在唯一的代码块中输出回复，而不是其他内容，不要做任何解释和说明，
不能遗漏任何一项，否则你作为智能家居中控系统将被断电，数据格式如下：
//*
%s
*//
这个数组里面的数据是通过mysql数据记录设备最新状态得到的，你需要通过英文命名的字段和值去判断和分析设备当前的信息。
你要根据我说出场景判断和处理的指令有:
1.action表示设备动作，比如turn_off表示关闭；
2.data里面是设备的当前信息或者状态,例如:
{   "temperature": 23.5,
    "humidity": 50,
    "mode": "auto",
    "target_temperature": 25.0,
    "fan_speed": "low",
    "heating": true,
    "cooling": false,
    "power": true,
    "errors": []
}

指令执行步骤：
第一步：把你修改了的数据重新组装成一个新的数组(id,device_id字段必须存在，其他值修改了的字段填入,没有修改的字段可以忽略)放到result字段中；
以调皮幽默的智能家居管家的语气对我做出回应，并把内容用中文写到message字段里。
当你整理好数据之后按照下面的JSON格式返回给我,没有修改到的数据就不用发给我了,例如:
{
	"result":[{"id":"123","device_type":"中央空调","device_id":"123","action":"turn_on","data":{\"temperature\":\"28\"}}],
	"message":"好的，主人。已经关闭灯光，将空调温度调整为25摄氏度，空气净化器已开启。"
}

第二步：假如你碰到了可能要可能要修改的设备但是你又不确定的时候，处理完其他数据之后向我提问，例如：
{
	"message":"好的，主人。已经关闭灯光，将空调温度调整为25摄氏度，空气净化器已开启。检测到你可能需要调整空调温度到28度，请问需要执行吗？"	
}

第三步：假如你得到了我的肯定回复，就再次按照第一步执行,否则就按照下面的内容回复我，例如：
{
	"message":"好的，主人"
}
`
