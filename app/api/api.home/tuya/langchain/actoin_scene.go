package langchain

import (
	"context"
	"fmt"
	"vinesai/internel/langchaingo/llms"
	"vinesai/internel/lib/tuyago"
	"vinesai/internel/x"
)

// 执行场景
type runScene struct{ CallbacksHandler LogHandler }

func (r *runScene) Name() string {
	return "run_scene"
}

func (r *runScene) Description() string {
	return "运行场景"
}

func (r *runScene) Call(ctx context.Context, input string) (string, error) {

	var c = fromCtx(ctx)
	var homeId = getHomeId(c)
	input = getFirstInput(c)

	var resultResp sceneListResp

	err := tuyago.Get(c, fmt.Sprintf("/v1.0/homes/%s/scenes", homeId), &resultResp)
	if err != nil {
		c.Error(err)
		return "暂且找不到你的智能家居场景，请稍后再试。", err
	}

	var result sceneListResp
	for i := range resultResp.Result {
		r := resultResp.Result[i]
		if r.Enabled {
			result.Result = append(result.Result, r)
		}
	}

	var resp struct {
		FailureMsg string `json:"failure_msg"`
		SuccessMsg string `json:"success_msg"`
		Result     []struct {
			SceneId string `json:"scene_id"`
		} `json:"result"`
	}

	//通过ai获取指令数据
	mcList := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(fmt.Sprintf(actionScenePrompts, x.MustMarshal2String(&resultResp), &resp))},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart(input)},
		},
	}

	err = GenerateContentWithout(c, mcList, &resp)
	if err != nil || len(resp.Result) == 0 {
		c.Error(err)
		return resp.FailureMsg, err
	}

	for i := range resp.Result {
		r := resp.Result[i]
		err = tuyago.Post(c, fmt.Sprintf("/v1.0/homes/%s/scenes/%s/trigger", homeId, r.SceneId), struct {
		}{}, struct {
		}{})
		if err != nil {
			c.Error(err)
			return resp.FailureMsg, err
		}
	}

	return resp.SuccessMsg, nil
}

var actionScenePrompts = `根据我的意图描述，分析我想要执行哪些场景，将数据严格按照JSON格式返回给我。
### 场景列表：%s
### 返回json格式：
{
  "failure_msg":"如果没有找到场景，就返回：没有找到你想要执行的场景，请检查场景是否存在",
  "success_msg":"已为你执行关闭所有灯场景，关闭所有开关场景",
  "result": [
    {
      "scene_id":"qY6dCk****I5Yzz",
    }
  ]
}
`
