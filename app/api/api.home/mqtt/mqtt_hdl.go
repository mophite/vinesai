package mqtt

import (
	"context"
	"fmt"
	"net/http"
	"vinesai/internel/ava"
	"vinesai/internel/config"
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

	toAI := fmt.Sprintf(botTmp, string(data))
	c.Debugf("TO |data=%s", toAI)

	resp, err := miniprogram.OpenAi.CreateCompletion(context.Background(), openai.CompletionRequest{
		Model:       openai.GPT3Dot5TurboInstruct,
		Prompt:      toAI,
		Temperature: config.GConfig.OpenAI.Temperature,
		TopP:        config.GConfig.OpenAI.TopP,
		MaxTokens:   1000,
	})

	if err != nil {
		c.Error(err)
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = x.StatusInternalServerError
		return
	}

	c.Debugf("FROM |data=%s", ava.MustMarshalString(&resp))

	var result struct {
		Result  string `json:"result"`
		Message string `json:"message"`
	}

	ava.MustUnmarshal(ava.MustMarshal(&resp), &result)

	var device []*db_hub.Device
	ava.MustUnmarshal(ava.StringToBytes(result.Result), device)

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
