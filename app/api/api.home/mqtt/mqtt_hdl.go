package mqtt

import (
	"context"
	"fmt"
	"net/http"
	"vinesai/internel/ava"
	"vinesai/internel/db"
	"vinesai/internel/db/db_hub"
	"vinesai/internel/x"
	"vinesai/proto/pmini"

	"github.com/sashabaranov/go-openai"
	"vinesai/app/api/api.home/miniprogram"
)

// 用户向云后台发送指令
type MqttHub struct {
}

func (m *MqttHub) Order(c *ava.Context, req *pmini.OrderReq, rsp *pmini.OrderRsp) {

	if req.Content == "" || req.UserId == "" {
		rsp.Code = http.StatusBadRequest
		rsp.Msg = "你想要说什么"
		return
	}

	//根据用户id查询出用户的所有设备
	var deviceList []*db_hub.Device
	err := db.
		GMysql.
		Table(db_hub.TableDeviceList).
		Where("user_id=?", req.UserId).
		Order("created_at desc").
		Limit(3).
		Find(&deviceList).Error
	if err != nil {
		c.Error(err)
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = x.StatusInternalServerError
		return
	}

	//todo 取出三次对话的记录同时发送给AI，要建立一个会话记录表

	//将指令发给AI
	data := ava.MustMarshal(deviceList)
	c.Debugf("content=%s |deviceList=%s", req.Content, string(data))

	toAI := fmt.Sprintf(botTmp, string(data), req.Content)

	//resp, err := miniprogram.OpenAi.CreateCompletion(context.Background(), openai.CompletionRequest{
	//	Model:       openai.GPT3Dot5TurboInstruct,
	//	Prompt:      toAI,
	//	Temperature: config.GConfig.OpenAI.Temperature,
	//	TopP:        config.GConfig.OpenAI.TopP,
	//	MaxTokens:   1000,
	//})

	var top = openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: toAI,
	}

	var next = openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: req.Content,
	}

	var msgList = make([]openai.ChatCompletionMessage, 0, 3)
	msgList = append(msgList, top, next)

	c.Debugf("TO |data=%s", ava.MustMarshalString(msgList))

	resp, err := miniprogram.OpenAi.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo,
			Messages: msgList,
			//Temperature: config.GConfig.OpenAI.Temperature,
			//TopP:        config.GConfig.OpenAI.TopP,
			Temperature: 0.5,
			TopP:        1,
			N:           1,
			MaxTokens:   4000,
		},
	)

	if err != nil {
		c.Error(err)
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = x.StatusInternalServerError
		return
	}

	c.Debugf("FROM |data=%s", ava.MustMarshalString(&resp))

	if len(resp.Choices) == 0 {
		c.Error(err)
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = x.StatusInternalServerError
		return
	}

	var result struct {
		Result  []*db_hub.Device `json:"result"`
		Message string           `json:"message"`
	}

	ava.MustUnmarshal(ava.StringToBytes(resp.Choices[0].Message.Content), &result)

	c.Debugf("text=%s |result=%v", resp.Choices[0].Message.Content, ava.MustMarshalString(&result))

	device := result.Result
	//更新设备状态,并向设备发送推送
	for i := range device {
		var d db_hub.Device
		d.Action = device[i].Action
		d.Data = device[i].Data
		err = db.GMysql.Table(db_hub.TableDeviceList).Where("id=?", device[i].ID).Updates(d).Error
		if err != nil {
			c.Error(err)
			rsp.Code = http.StatusInternalServerError
			rsp.Msg = x.StatusInternalServerError
			return
		}

		//向设备发送推送
		mqttPublish(device[i].DeviceId, req.UserId, ava.MustMarshalString(device[i]))
	}

	rsp.Code = http.StatusOK
	rsp.Msg = result.Message

}
