package langchain

import (
	"context"
	"fmt"
	"vinesai/internel/lib/tuyago"
	"vinesai/internel/x"
)

// 查询家庭场景
type autoQuery struct{ CallbacksHandler LogHandler }

func (s *autoQuery) Name() string {
	return "query_home_auto"
}

func (s *autoQuery) Description() string {
	return "查询智能家居自动化"
}

func (s *autoQuery) Call(ctx context.Context, input string) (string, error) {
	var c = fromCtx(ctx)
	var homeId = getHomeId(c)
	input = getFirstInput(c)

	var resultResp struct {
		Result []struct {
			Name    string `json:"name"`
			Enabled bool   `json:"enabled"`
		} `json:"result"`
	}

	err := tuyago.Get(c, fmt.Sprintf("/v1.0/homes/%s/automations", homeId), &resultResp)
	if err != nil {
		c.Error(err)
		return "暂且找不到你的智能家居自动化，请稍后再试。", err
	}

	resp, err := GenerateContentTurbo(c, fmt.Sprintf(queryAutoPrompts, x.MustMarshal2String(resultResp)), input)
	if err != nil {
		c.Error(err)
		return "我找不到你的自动化数据啦，需要我为你再找找看吗", err
	}

	return resp, nil

}

var queryAutoPrompts = `根据我的意图描述，告诉我智能家居自动化相关的信息，严格按照json格式返回。
### 自动化列表：%s
说明：enabled：false表示自动化是禁用状态，true表示自动化是启用中。
### 用俏皮人性化的语气回复我，例如：你有3个自动化，启用中2个，禁用状态1个。

当前对话：
{{.history}}
Human: {{.input}}
AI:`