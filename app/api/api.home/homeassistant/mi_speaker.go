package homeassistant

import (
	"strings"
	"sync/atomic"
	"time"
	"vinesai/internel/ava"
	"vinesai/internel/x"
)

type entity struct {
	home     string
	entityId string
}

// 判断是不是消息音响实体
func isXiaoMiSpeaker(entityId string) bool {
	return strings.HasPrefix(entityId, "sensor.xiaomi_") && strings.HasSuffix(entityId, "_conversation")
}

func isHumanBodySensor(deviceClass string) bool {
	return deviceClass == "motion"

}

// 启动所有小爱音响的静音
func runXiaoMiSpeaker(home string) {

	//初始化的时候就要执行，获取所有实体
	_, entities, err := getStates(ava.Background(), home)
	if err != nil {
		ava.Error(err)
		return
	}

	//判断是否是音响实体,例如:sensor.xiaomi_l05c_7e78_conversation
	for i := range entities {
		id := entities[i].EntityId

		if isXiaoMiSpeaker(id) {
			ava.Debug(id)

			gHub.addEntity(home, &entity{home: home, entityId: id})
		}
	}
}

// 接收到最新的语音消息，发给AI
// ai返回的内容让小爱播放出来
func receiveMessage(c *ava.Context, home, message string, isSend bool) error {

	result, err := SelectAI(c, "123", message)
	if err != nil {
		c.Error(err)
		return err
	}

	id := time.Now().UnixNano()

	for i := range result.Commands {
		command := result.Commands[i]
		//发起设备控制
		callServiceWs(atomic.AddInt64(&id, 1), home, command)
	}

	//发送message给小爱音响
	if isSend {
		send2XiaomiSpeaker(c, home, result.Reply)
	}
	return nil
}

func send2XiaomiSpeaker(c *ava.Context, home, message string) {
	//关闭小爱的静音
	e := gHub.getEntity(home)
	if e == nil {
		return
	}

	//发送语音消息给小爱音响
	conn := gHub.getConn(home)

	var data = wsCallService{
		Id:      time.Now().UnixNano(),
		Type:    "call_service",
		Domain:  "xiaomi_miot",
		Service: "intelligent_speaker",
	}
	data.Target.EntityId = e.entityId
	data.ServiceData = struct {
		Text    string `json:"text"`
		Execute bool   `json:"execute"`
		Silent  bool   `json:"silent"`
	}{message, false, false}

	c.Debugf("send2XiaomiSpeaker |data=%s", x.MustMarshal2String(&data))
	conn.WriteJSON(&data)

}
