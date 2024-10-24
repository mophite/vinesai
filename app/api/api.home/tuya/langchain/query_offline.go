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

type queryOffline struct{ CallbacksHandler LogHandler }

func (q *queryOffline) Name() string {
	return "query_offline_device"
}

func (q *queryOffline) Description() string {
	return "查询离线设备"
}

func (q *queryOffline) Call(ctx context.Context, input string) (string, error) {

	var c = fromCtx(ctx)
	var homeId = getHomeId(c)
	input = getFirstInput(c)

	var msg = "请告诉我你要查询什么"

	var devices []*queryOnlineOrOfflineData
	var filter = bson.M{"homeid": homeId, "online": false}
	cur, err := db.Mgo.Collection(mgoCollectionDevice).Find(context.Background(), filter)
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

	mcList := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(fmt.Sprintf(queryOfflinePrompts, x.MustMarshal2String(devices)))},
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

var queryOfflinePrompts = `请分析我的意图，并根据以下设备列表，严格按照JSON 格式返回。
### 输入：以下是所有离线的设备列表：%s
### 返回的 JSON 格式要求如下：
{
	"content": "根据内容用人性化的语言描述。例如：你的家里没有设备离线"
}
### 注意事项：
如果没有离线设备就告诉我：没有设备离线`
