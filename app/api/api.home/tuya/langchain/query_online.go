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

type queryOnline struct{ CallbacksHandler LogHandler }

func (q *queryOnline) Name() string {
	return "query_online_device"
}

func (q *queryOnline) Description() string {
	return "查询在线设备"
}

func (q *queryOnline) Call(ctx context.Context, input string) (string, error) {

	var c = fromCtx(ctx)
	var homeId = getHomeId(c)
	input = getFirstInput(c)

	var msg = "请告诉我你要查询什么"

	var devices []*queryOnlineOrOfflineData
	var filter = bson.M{"homeid": homeId, "online": true}
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

	mcList := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(fmt.Sprintf(queryOnlinePrompts, x.MustMarshal2String(devices)))},
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

var queryOnlinePrompts = `分析我的意图，并根据以下设备列表返回json数据结果。
### 所有在线的设备列表：%s
### 返回json格式和例子：
{
	"content":"你在客厅有两个设备"
}`
