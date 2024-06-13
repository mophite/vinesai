package mqtt

import (
	"fmt"
	"strings"
	"time"
	"vinesai/internel/ava"
	"vinesai/internel/db"
	"vinesai/internel/db/db_hub"
	"vinesai/internel/x"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"gorm.io/gorm"
)

// 插座
var deviceReportTopic = "home/report/device"
var deviceControlTopic = "home/control/device/%s/%s" //独立设备id的topic

// 红外设备
var deviceReportInfrared = "v2/tele/wir1/#"
var deviceControlInfrared = "v2/cmnd/wir1/%s/IR/irsend" //user_id,device_id

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
	go mqttReportSubscribeInfrared()

	return nil
}

var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("主题: %s\n", msg.Topic())
	fmt.Printf("信息: %s\n", string(msg.Payload()))
}

// 发布消息给指定的客户端
func mqttPublish(delayTime int, deviceId, userId, data string) {

	var topic = fmt.Sprintf(deviceControlTopic, userId, deviceId)
	ava.Debugf("mqttPublish |topic=%s |data=%s |delay_time=%v", topic, data, delayTime)

	f := func() {

		//retained false服务器不保留消息
		token := client.Publish(topic, 0, false, data)
		token.Wait()
		if token.Error() != nil {
			ava.Debugf("mqttPublish failure |deviceId=%s |err=%v", deviceId, token.Error())
			ava.Error(token.Error())
		}
	}

	//延时执行
	if delayTime > 0 {
		x.TimingwheelAfter(time.Second*time.Duration(delayTime), f)
		return
	}

	f()

}

// 发布消息给指定的客户端,红外设备
func mqttPublishInfrared(delayTime int, deviceId, userId, data string) {

	var topic = fmt.Sprintf(deviceControlInfrared, deviceId)
	ava.Debugf("mqttPublishInfrared |topic=%s |data=%s |delay_time=%v", topic, data, delayTime)

	f := func() {
		//retained false服务器不保留消息
		token := client.Publish(topic, 0, false, data)
		//todo 保证消息推送成功
		token.Wait()
		if token.Error() != nil {
			ava.Debugf("mqttPublishInfrared failure |deviceId=%s |err=%v", deviceId, token.Error())
			ava.Error(token.Error())
		}
	}

	if delayTime > 0 {
		x.TimingwheelAfter(time.Second*time.Duration(delayTime), f)
		return
	}

	f()

}

// 红外设备
func mqttReportSubscribeInfrared() {
	token := client.Subscribe(deviceReportInfrared, byte(0), func(c mqtt.Client, message mqtt.Message) {

		/*

			[DBUG] mqtt_biz.go:95 2024-05-15T23:22:01.0990 subscribe topic=home/report/device|payload=Online |topic=v2/tele/wir1/846A44/LWT
			[DBUG] mqtt_biz.go:95 2024-05-15T23:22:01.0991 subscribe topic=home/report/device|payload={"Info1":{"Module":"ESP32C3","Version":"12.4.0.11(USTONE-HA)","FallbackTopic":"cmnd/wir1_846A44_fb/","GroupTopic":"v2/cmnd/wir1/all/"}} |topic=v2/tele/wir1/846A44/INFO1
			[DBUG] mqtt_biz.go:95 2024-05-15T23:22:01.0991 subscribe topic=home/report/device|payload={"Info2":{"WebServerMode":"Admin","Hostname":"wir1-846A44-2628","IPAddress":"192.168.1.70","IP6Global":"2408:826a:22:9ee6:9e9e:6eff:fe84:6a44","IP6Local":"fe80::9e9e:6eff:fe84:6a44"}} |topic=v2/tele/wir1/846A44/INFO2
			[DBUG] mqtt_biz.go:95 2024-05-15T23:22:01.0992 subscribe topic=home/report/device|payload={"Info3":{"RestartReason":"Vbat power on reset","BootCount":9}} |topic=v2/tele/wir1/846A44/INFO3
			[DBUG] mqtt_biz.go:95 2024-05-15T23:22:05.5806 subscribe topic=home/report/device|payload={"Time":"2024-05-15T23:22:04","Uptime":"0T00:00:11","UptimeSec":11,"Heap":216,"SleepMode":"Dynamic","Sleep":50,"LoadAvg":22,"MqttCount":1,"Wifi":{"AP":1,"SSId":"CU_muV8","BSSId":"08:79:8C:DE:EE:0C","Channel":11,"Mode":"11n","RSSI":100,"Signal":-44,"LinkCount":1,"Downtime":"0T00:00:06"}} |topic=v2/tele/wir1/846A44/STATE
		*/
		ava.Debugf("subscribe topic=%s|payload=%s |topic=%s", deviceReportTopic, string(message.Payload()), message.Topic())

		t := strings.Split(message.Topic(), "/")
		if len(t) != 5 {
			return
		}

		var (
			did  = t[3]
			info = t[4]
		)

		if info == "STATE" {
			//过滤心跳包
			return
		}

		if info != "LWT" && info != "RESULT" {
			//设备上线
			//设备控制结果
			return
		}

		//第一次连接
		err := db.GMysql.Transaction(func(tx *gorm.DB) error {

			//如果不存在设备则添加一个设备
			var d db_hub.Device
			err := tx.Table(db_hub.TableDeviceList).
				Where("device_id=? AND user_id=?", did, "123").
				Take(&d).
				Error

			if err != nil && err.Error() == gorm.ErrRecordNotFound.Error() {
				//数据不存在,插入数据
				device := db_hub.Device{
					DeviceType: "2",
					DeviceZn:   "红外控制器",
					DeviceEn:   "Infrared",
					DeviceID:   did,
					DeviceDes:  "红外控制设备，请编辑描述",
					Version:    "0.0.1",
					UserID:     "123",
					Control:    "1",
					Ip:         "",
					Wifi:       "",
				}
				return tx.Table(db_hub.TableDeviceList).Create(&device).Error
			}

			//如果是设备上线只需要添加设备
			if info == "LWT" {
				return nil
			}

			//如果设别存在则更新设备的开关状态
			if info == "RESULT" {
				type Result struct {
					IrReceived struct {
						Bits   int64  `json:"Bits"`
						Data   string `json:"Data"`
						Irhvac struct {
							Beep     string `json:"Beep"`
							Celsius  string `json:"Celsius"`
							Clean    string `json:"Clean"`
							Econo    string `json:"Econo"`
							FanSpeed string `json:"FanSpeed"`
							Filter   string `json:"Filter"`
							Light    string `json:"Light"`
							Mode     string `json:"Mode"`
							Model    int64  `json:"Model"`
							Power    string `json:"Power"`
							Quiet    string `json:"Quiet"`
							Sleep    int64  `json:"Sleep"`
							SwingH   string `json:"SwingH"`
							SwingV   string `json:"SwingV"`
							Temp     int64  `json:"Temp"`
							Turbo    string `json:"Turbo"`
							Vendor   string `json:"Vendor"`
						} `json:"IRHVAC"`
						Protocol    string  `json:"Protocol"`
						RawData     string  `json:"RawData"`
						RawDataInfo []int64 `json:"RawDataInfo"`
						Repeat      int64   `json:"Repeat"`
					} `json:"IrReceived"`
				}

				var result Result
				err = ava.Unmarshal(message.Payload(), &result)
				if err != nil {
					ava.Error(err)
					return err
				}
				if result.IrReceived.Irhvac.Vendor == "" {
					return nil
				}

				//存在则更新数据
				updates := make(map[string]interface{}, 10)
				//if device.DeviceZn != "" {
				//	updates["device_zn"] = device.DeviceZn
				//}
				//if device.DeviceEn != "" {
				//	updates["device_en"] = device.DeviceEn
				//}
				//if device.DeviceDes != "" {
				//	updates["device_des"] = device.DeviceDes
				//}
				if result.IrReceived.Irhvac.Power != "" {
					control := 0
					if result.IrReceived.Irhvac.Power == "On" {
						control = 2
					}

					if result.IrReceived.Irhvac.Power == "Off" {
						control = 1
					}

					updates["control"] = control
				}
				return db.GMysql.Table(db_hub.TableDeviceList).
					Where("device_id=? AND user_id=?", did, "123").
					Updates(updates).Error
			}
			return nil
		})

		if err != nil {
			ava.Error(err)
		}

	})

	if token.Wait() && token.Error() != nil {
		panic(fmt.Sprintf("sbscribe |topic=%s |err=%v", deviceReportTopic, token.Error()))
	}

}

// 测试版本智能插座
// 订阅客户端发送的消息:数据上报
// 将数据状态同步到数据库中
// todo 不同品牌的设备不同的topic
func mqttReportSubscribe() {
	token := client.Subscribe(deviceReportTopic, byte(0), func(c mqtt.Client, message mqtt.Message) {
		ava.Debugf("subscribe topic=%s|payload=%s", deviceReportTopic, string(message.Payload()))

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

			if err != nil && err.Error() == gorm.ErrRecordNotFound.Error() {
				//数据不存在,插入数据
				return tx.Table(db_hub.TableDeviceList).Create(&device).Error
			}

			//存在则更新数据
			updates := make(map[string]interface{}, 10)
			//if device.DeviceZn != "" {
			//	updates["device_zn"] = device.DeviceZn
			//}
			//if device.DeviceEn != "" {
			//	updates["device_en"] = device.DeviceEn
			//}
			//if device.DeviceDes != "" {
			//	updates["device_des"] = device.DeviceDes
			//}
			if device.Control != "" {
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

//var botTmp = `你现在是一个智能家居中控系统。根据我提供的设备数据,通过英文字段和值来分析设备的当前状态,并根据我描述的场景来控制它们。
//场景控制指令下达后,请严格按照以下JSON格式,在唯一的代码块中输出结果,不要有其他多余内容:
//{
//"result": [
//{
//"user_id": "123",
//"device_type": 1,
//"device_id": "8CCE4E522308",
//"control": 1
//}
//],
//"message": "您好,主人!卧室灯已经关闭。今晚祝您做个好梦~"
//}
//字段说明如下,未提及的字段直接忽略:
//device_type:设备类型
//device_zn:设备中文名称
//device_en:设备英文名称
//device_id:设备ID
//user_id:所属用户ID
//control:开关状态,1为断电,2为通电
//注意事项:
//
//1、user_id、device_id、device_type 是必须字段,不可修改它们的值,其他字段只保留有变更的。
//2、将处理后的设备数据放入 result 数组中。
//3、以幽默活泼的口吻回应场景执行结果,在 message 字段中说明调整的设备。
//
//我会提供设备数据如下:
//%s
//请分析以上数据,等待我的控制指令。一旦收到指令,即刻按要求输出JSON结果。`

var botTmp = `   
你是一个智能家居助手,可以根据用户描述的场景推断用户的需求,并输出相应的控制命令。
[设备清单]信息如下:
%s

控制命令严格以下面 JSON 格式输出,开头和末尾后不要有多余的内容，否则被视为非法数据，
如果场景涉及多条指令,请将所有指令放在一个名为 "commands" 的数组中,例如：
{
    "voice": "风趣幽默的智能音箱语气回复的文本内容",
    "commands": [
        {"user_id": "设备清单中的用户ID", "device_type": "设备清单设备类型1","device_id": "设备清单设备ID1", "control": "开关控制，1关，2开","delay_time":"延时执行执行时间，例如：5表示5秒"},
        {"user_id": "设备清单中的用户ID", "device_type": "设备清单设备类型2","device_id": "设备清单设备ID1", "control": "开关控制，1关，2开","delay_time":"延时执行执行时间，例如：5表示5秒"}
    ]
}

可生成的命令类型包括:
1、关闭设备: {"user_id": "设备清单中的用户ID", "device_type": "设备清单设备类型","device_id": "设备清单设备ID1", "control": "1","delay_time":"10"}
2、打开设备: {"user_id": "设备清单中的用户ID", "device_type": "设备清单设备类型","device_id": "设备清单设备ID1", "control": "2","delay_time":"0"}
3、延时10秒执行,其中delay(1表示延时操作，0表示非延时),delay_time表示延时多少秒执行:
{"user_id": "设备清单中的用户ID", "device_type": "设备清单设备类型1","device_id": "设备清单设备ID1", "control": "1","delay_time":"10"}


如果无法根据用户的描述生成合适的控制命令,请回复 "无法生成控制命令"。`
