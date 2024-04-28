package mqtt

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"vinesai/internel/ava"
	"vinesai/internel/db"
	"vinesai/internel/db/db_hub"
	"vinesai/internel/x"
	"vinesai/proto/pmini"

	"github.com/sashabaranov/go-openai"
	"go.mongodb.org/mongo-driver/bson"
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
	// 定义查询条件
	filter := bson.M{"user_id": req.UserId}

	collection := db.GMongo.Database(db_hub.DatabaseMongoVinesai).Collection(db_hub.CollectionDevice)

	// 执行查询操作
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		c.Error(err)
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = x.StatusInternalServerError
		return
	}

	defer cursor.Close(context.Background())

	// 遍历查询结果
	var devices []*db_hub.Device
	for cursor.Next(context.Background()) {
		var device *db_hub.Device
		if err := cursor.Decode(&device); err != nil {
			c.Error(err)
			continue
		}

		devices = append(devices, device)
	}

	// 检查遍历过程中是否出错
	if err := cursor.Err(); err != nil {
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
		Role:    openai.ChatMessageRoleUser,
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

	str := resp.Choices[0].Message.Content
	str = strings.ReplaceAll(str, "\n", "")
	str = strings.ReplaceAll(str, "\t", "")
	str = strings.ReplaceAll(str, `\`, "")

	ava.MustUnmarshal(ava.StringToBytes(str), &result)

	c.Debugf("text=%s |result=%v", str, ava.MustMarshalString(&result))

	//更新设备状态,并向设备发送推送
	for i := range result.Result {
		var d = result.Result[i]
		//// 定义更新条件
		//filterDeviceId := bson.M{"user_id": d.UserID, "device_id": d.DeviceID}
		//updateMap := bson.M{
		//	"updated_at": time.Now().UnixMilli(),
		//}
		////判断开关
		//if d.Switch != 0 {
		//	updateMap["switch"] = d.Switch
		//}
		////todo 其他判断
		//
		//// 定义更新内容
		//update := bson.M{"$set": updateMap}
		//
		//// 执行更新操作
		//_, err = collection.UpdateMany(context.Background(), filterDeviceId, update)
		//if err != nil {
		//	c.Error(err)
		//	rsp.Code = http.StatusInternalServerError
		//	rsp.Msg = x.StatusInternalServerError
		//	return
		//}

		//向设备发送推送
		//发送推送的时候要做转换处理
		toDevice, err := db_hub.Device2Adaptor(d)
		if err != nil {
			c.Error(err)
			rsp.Code = http.StatusInternalServerError
			rsp.Msg = x.StatusInternalServerError
			return
		}
		mqttPublish(c, d.DeviceID, req.UserId, ava.MustMarshalString(toDevice))
	}

	rsp.Code = http.StatusOK
	rsp.Msg = result.Message
	order := pmini.OrderData{
		Order:  req.Content,
		ToAi:   ava.MustMarshalString(msgList),
		FromAi: ava.MustMarshalString(&resp),
	}
	rsp.Data = &order

}
