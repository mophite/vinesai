package langchain

import (
	"context"
	"fmt"
	"vinesai/internel/ava"
	"vinesai/internel/db"
	"vinesai/internel/x"

	"vinesai/internel/langchaingo/llms"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type queryDevice struct{ CallbacksHandler LogHandler }

func (q *queryDevice) Name() string {
	return "query_device"
}

func (q *queryDevice) Description() string {
	return "查询设备状态，设备总数等；非在线，离线关键字"
}

func (q *queryDevice) Call(ctx context.Context, input string) (string, error) {

	var c = fromCtx(ctx)
	var homeId = getHomeId(c)
	input = getFirstInput(c)

	var msg = "请告诉我你要查询什么"

	var devices []*queryDevicesData
	var filter = bson.M{"homeid": homeId}
	cur, err := db.Mgo.Collection(mgoCollectionNameDevice).Find(context.Background(), filter)
	if err != nil {
		ava.Error(err)
		return "服务器出小毛病了", err
	}

	defer cur.Close(ctx)

	err = cur.All(context.Background(), &devices)
	if err != nil {
		ava.Error(err)
		return "服务器出小毛病了", err
	}

	//todo 分两步，先问ai需要获取哪些设备，再查结果
	mcList := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(fmt.Sprintf(queryDevicePrompts, x.MustMarshal2String(devices)))},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart(input)},
		},
	}

	var req queryOnlineResp
	err = GenerateContentWithout(c, mcList, &req)
	if err != nil {
		c.Error(err)
		return msg, err
	}

	msg = req.Content

	return msg, nil
}

var queryDevicePrompts = `分析我的意图，并根据以下设备列表，用最严格按照JSON格式返回结果。
### 输入：所有设备列表：%s
### 返回JSON格式：
{
	"content":"客厅灯亮度500"
}

### 注意事项：
1.根据Name字段的数量统计设备数量；
2.如果是查询设备状态，根据Status总结设备的重要信息
3.如果需要告诉我的详细设备信息数超过5个，直接总结告诉我`
