package homeassistant

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"
	"vinesai/internel/ava"
	"vinesai/internel/x"

	"github.com/RussellLuo/timingwheel"
	"github.com/sashabaranov/go-openai"
	"vinesai/app/api/api.home/miniprogram"
)

type serviceData struct {
	ID          int64       `json:"id"`
	Type        string      `json:"type"`
	Domain      string      `json:"domain"`
	Service     string      `json:"service"`
	ServiceData interface{} `json:"service_data"`
	Target      struct {
		EntityId string `json:"entity_id"`
	} `json:"target"`
}

type entity struct {
	timer    *timingwheel.Timer
	home     string
	entityId string
}

// 当ai返回内容之后，暂停轮询
func (e *entity) watchAndStop() {
	e.timer.Stop()
}

// 轮询每秒向小爱发送空值，使它不说话
func (e *entity) rotate2Speaker() {

	conn := gHub.getConn(e.home)
	if conn == nil {
		ava.Debug("no such conn")
		return
	}

	var now = time.Now().UnixNano()

	var toBan = serviceData{
		ID:      now,
		Type:    "call_service",
		Domain:  "xiaomi_miot",
		Service: "intelligent_speaker",
	}
	toBan.Target.EntityId = e.entityId
	toBan.ServiceData = struct {
		Text    string `json:"text"`
		Execute bool   `json:"execute"`
		Silent  bool   `json:"silent"`
	}{"_", false, false}

	e.timer = x.TimingwheelTicker(time.Second, func() {
		fmt.Println("-------", x.MustMarshal2String(&toBan))
		//每隔2秒发送禁言
		conn.WriteJSON(&toBan)
	})
}

// 判断是不是消息音响实体
func isXiaoMiSpeaker(entityId string) bool {
	return strings.HasPrefix(entityId, "media_player.xiaomi_") && strings.HasSuffix(entityId, "_control")
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

			var e = &entity{home: home, entityId: id}
			e.rotate2Speaker()

			gHub.addEntity(home, e)
		}
	}
}

// 接收到最新的语音消息，发给AI
// ai返回的内容让小爱播放出来
func recevieMessage(home, message string) error {
	ava.Debug("--------event_change", message)
	//获取指令集
	c := ava.Background()
	services, err := getServices(c, home)
	if err != nil {
		ava.Error(err)
		return err
	}

	states, _, err := getStates(c, "123")
	if err != nil {
		c.Error(err)
		return err
	}

	//告诉ai指令
	var top = openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: fmt.Sprintf(aiTmp2Ws, services, states),
	}

	var next = openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message,
	}

	var msgList = make([]openai.ChatCompletionMessage, 0, 3)
	msgList = append(msgList, top, next)

	resp, err := miniprogram.OpenAi.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			//Model:    openai.GPT3Dot5Turbo,
			//Model: "claude-3-haiku-20240307",
			Model:    "gpt-4o",
			Messages: msgList,
			//Temperature: config.GConfig.OpenAI.Temperature,
			//TopP:        config.GConfig.OpenAI.TopP,
			Temperature:    0.5,
			TopP:           1,
			N:              1,
			MaxTokens:      4000,
			ResponseFormat: &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject},
		},
	)

	if err != nil {
		c.Error(err)
		return err
	}

	c.Debugf("FROM |data=%s", ava.MustMarshalString(&resp))

	if len(resp.Choices) == 0 {
		c.Error(err)
		return err
	}

	var result struct {
		Command []struct {
			ID          int64       `json:"id"`
			Type        string      `json:"type"`
			Domain      string      `json:"domain"`
			Service     string      `json:"service"`
			ServiceData interface{} `json:"service_data"`
			Target      interface{} `json:"target"`
		} `json:"serviceData"`
		Message string `json:"message"`
	}

	str := resp.Choices[0].Message.Content
	str = strings.ReplaceAll(str, "\n", "")
	str = strings.ReplaceAll(str, "\t", "")
	str = strings.ReplaceAll(str, `\`, "")

	ava.MustUnmarshal(ava.StringToBytes(str), &result)

	c.Debugf("text=%s |result=%v", str, ava.MustMarshalString(result))

	id := time.Now().UnixNano()

	for i := range result.Command {
		command := result.Command[i]
		command.ID = atomic.AddInt64(&id, 1)
		//发起设备控制
		callServiceWs(home, command)
	}

	//发送message给小爱音响
	send2XiaomiSpeaker(home, result.Message)
	return nil
}

func send2XiaomiSpeaker(home, message string) {
	//关闭小爱的静音
	e := gHub.getEntity(home)
	e.watchAndStop()

	//发送语音消息给小爱音响
	conn := gHub.getConn(home)

	var now = time.Now().UnixNano()

	var data = serviceData{
		ID:      now,
		Type:    "call_service",
		Domain:  "xiaomi_miot",
		Service: "intelligent_speaker",
	}
	data.Target.EntityId = e.entityId
	data.ServiceData = struct {
		Text    string `json:"text"`
		Execute bool   `json:"execute"`
		Silent  bool   `json:"silent"`
	}{message, false, true}

	conn.WriteJSON(&data)

	x.TimingwheelAfter(time.Second*5, func() {
		e.rotate2Speaker()
	})

}
