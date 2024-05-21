package mqtt

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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

func (m *MqttHub) DeviceEdit(c *ava.Context, req *pmini.DeviceEditReq, rsp *pmini.DeviceEditRsp) {
	if req.UserId == "" || req.DeviceId == "" {
		rsp.Code = http.StatusBadRequest
		rsp.Msg = "请求参数错误"
		return
	}

	if req.DeviceDes == "" && req.DeviceEn == "" && req.DeviceZn == "" {
		rsp.Code = http.StatusBadRequest
		rsp.Msg = "请输入要修改的数据"
		return
	}

	var updates = make(map[string]interface{}, 10)
	if req.DeviceEn != "" {
		updates["device_en"] = req.DeviceZn
	}

	if req.DeviceZn != "" {
		updates["device_zn"] = req.DeviceZn
	}

	if req.DeviceDes != "" {
		updates["device_des"] = req.DeviceDes
	}

	err := db.GMysql.
		Table(db_hub.TableDeviceList).
		Where("user_id=? AND device_id=?", req.UserId, req.DeviceId).
		Updates(updates).Error

	if err != nil {
		c.Error(err)
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = "更新数据失败"
		return
	}

	rsp.Code = http.StatusOK
	rsp.Msg = "设备描述已更新"

}

func (m *MqttHub) DeviceList(c *ava.Context, req *pmini.DeviceListReq, rsp *pmini.DeviceListRsp) {
	if req.UserId == "" {
		rsp.Code = http.StatusBadRequest
		rsp.Msg = "user_id不能为空"
		return
	}

	var devices []*db_hub.Device
	err := db.
		GMysql.
		Table(db_hub.TableDeviceList).
		Where("user_id=?", req.UserId).
		Order("created_at desc").
		Limit(100).
		Find(&devices).Error
	if err != nil {
		c.Error(err)
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = x.StatusInternalServerError
		return
	}

	c.Debugf("result=%s", ava.MustMarshalString(devices))

	rsp.Code = http.StatusOK
	rsp.Data = ava.MustMarshalString(devices)
}

func (m *MqttHub) Order(c *ava.Context, req *pmini.OrderReq, rsp *pmini.OrderRsp) {

	if req.Content == "" || req.UserId == "" {
		rsp.Code = http.StatusBadRequest
		rsp.Msg = "你想要说什么"
		return
	}

	//根据用户id查询出用户的所有设备
	var devices []*db_hub.Device
	err := db.
		GMysql.
		Table(db_hub.TableDeviceList).
		Where("user_id=?", req.UserId).
		Order("created_at desc").
		Limit(100).
		Find(&devices).Error
	if err != nil {
		c.Error(err)
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = x.StatusInternalServerError
		return
	}

	//如果不存在设备，则返回提醒用户
	if len(devices) == 0 {
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = "你还没有配对任何设备"
		return
	}

	//todo 取出三次对话的记录同时发送给AI，要建立一个会话记录表

	//将指令发给AI
	data := ava.MustMarshal(devices)
	c.Debugf("content=%s |devices=%s", req.Content, string(data))

	toAI := fmt.Sprintf(botTmp, string(data))

	var top = openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: toAI,
	}

	var next = openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: req.Content,
	}

	var msgList = make([]openai.ChatCompletionMessage, 0, 3)
	msgList = append(msgList, top, next)

	c.Debugf("TO |data=%s", ava.MustMarshalString(msgList))

	resp, err := miniprogram.OpenAi.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			//Model:    openai.GPT3Dot5Turbo,
			//Model:    "claude-3-haiku-20240307",
			Model:    openai.GPT3Dot5Turbo,
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
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = "我有点晕。"
		return
	}

	c.Debugf("FROM |data=%s", ava.MustMarshalString(&resp))

	if len(resp.Choices) == 0 {
		c.Error(err)
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = "我犯迷糊了。"
		return
	}

	var result struct {
		Commands []*db_hub.Device `json:"commands"`
		Voice    string           `json:"voice"`
	}

	str := resp.Choices[0].Message.Content
	str = strings.ReplaceAll(str, "\n", "")
	str = strings.ReplaceAll(str, "\t", "")
	str = strings.ReplaceAll(str, `\`, "")
	//str = x.AiResultRex.FindString(str)

	ava.MustUnmarshal(ava.StringToBytes(str), &result)

	c.Debugf("text=%s |result=%v", str, ava.MustMarshalString(&result))

	//更新设备状态,并向设备发送推送
	for i := range result.Commands {
		//是否延时推送
		var d = result.Commands[i]
		//向设备发送推送
		//发送推送的时候要做转换处理
		switch d.DeviceType {
		case "1":
			key, err := strconv.Atoi(d.Control)
			if err != nil {
				c.Error(err)
				continue
			}
			toDevice := &db_hub.SocketMiniV2{
				Type: "event",
				Key:  key - 1,
			}
			delayTime, err := strconv.Atoi(d.DelayTime)
			if err != nil {
				c.Error(err)
				continue
			}

			mqttPublish(delayTime, d.DeviceID, req.UserId, ava.MustMarshalString(toDevice))

		case "2":
			power := ""
			if d.Control == "1" {
				power = "Off"
			}
			if d.Control == "2" {
				power = "On"
			}

			toDevice := &db_hub.Infrared{Power: power}

			delayTime, err := strconv.Atoi(d.DelayTime)
			if err != nil {
				c.Error(err)
				continue
			}

			mqttPublishInfrared(delayTime, d.DeviceID, req.UserId, ava.MustMarshalString(toDevice))
		}
	}

	rsp.Code = http.StatusOK
	rsp.Msg = result.Voice
	order := pmini.OrderData{
		Order:  req.Content,
		ToAi:   ava.MustMarshalString(msgList),
		FromAi: ava.MustMarshalString(&resp),
	}
	rsp.Data = &order
}
