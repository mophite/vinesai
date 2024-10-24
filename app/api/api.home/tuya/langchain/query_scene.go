package langchain

import (
	"context"
	"fmt"
	"vinesai/internel/db"
	"vinesai/internel/lib/tuyago"
	"vinesai/internel/x"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// 查询家庭场景
type sceneQuery struct{ CallbacksHandler LogHandler }

func (s *sceneQuery) Name() string {
	return "query_home_scene"
}

func (s *sceneQuery) Description() string {
	return "查询智能家居场景"
}

func (s *sceneQuery) Call(ctx context.Context, input string) (string, error) {
	var c = fromCtx(ctx)
	var homeId = getHomeId(c)
	input = getFirstInput(c)

	var resultResp struct {
		Result []struct {
			Name    string         `json:"name"`
			Enabled bool           `json:"enabled"`
			Actions []actions4Name `json:"actions"`
		} `json:"result"`
	}

	err := tuyago.Get(c, fmt.Sprintf("/v1.0/homes/%s/scenes", homeId), &resultResp)
	if err != nil {
		c.Error(err)
		return "暂且找不到你的智能家居场景，请稍后再试。", err
	}

	for i := range resultResp.Result {
		data := resultResp.Result[i].Actions
		for ii := range data {
			var d mgoDocDevice
			err = db.Mgo.Collection(mgoCollectionNameDevice).FindOne(context.Background(), bson.M{"_id": data[ii].EntityID}).Decode(&d)
			if err != nil {
				c.Error(err)
				continue
			}
			resultResp.Result[i].Actions[ii].Name = d.Name
		}
	}

	resp, err := GenerateContentTurbo(c, fmt.Sprintf(queryScenePrompts, x.MustMarshal2String(resultResp)), input)
	if err != nil {
		c.Error(err)
		return "我找不到你的场景数据啦，需要我为你再找找看吗", err
	}

	return resp, nil
}

var queryScenePrompts = `根据我的意图描述，告诉我智能家居场景相关的信息。
### 场景列表：%s
说明：enabled：false表示场景是禁用状态，true表示场景是启用中。
### 用俏皮人性化的语气回复我，不要返回无关的信息，例如：你有3个场景，启用中2个，禁用状态1个。

当前对话：
{{.history}}
Human: {{.input}}
AI:`
