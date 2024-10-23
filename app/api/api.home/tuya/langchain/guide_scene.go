package langchain

import (
	"context"
	"errors"
	"fmt"
	"vinesai/internel/db"

	"github.com/redis/go-redis/v9"
)

type guideScene struct{ CallbacksHandler LogHandler }

func (g *guideScene) Name() string {
	return "smart_home_scene_setting_guide"
}

func (g *guideScene) Description() string {
	return "根据用户意图，正常对话；然后分析用户意图描述，如果有需要，引导用户设置某种场景"
}

func (g *guideScene) Call(ctx context.Context, input string) (string, error) {
	var c = fromCtx(ctx)
	input = getFirstInput(c)

	devicesNameResult, err := db.GRedis.Get(context.Background(), redisKeyTuyaSummaryDeviceName+getHomeId(c)).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		c.Error(err)
		return "我开小差了", err
	}

	result, err := GenerateContentTurbo(c, fmt.Sprintf(guideScenePrompts, devicesNameResult), input)
	if err != nil {
		c.Error(err)
		if result != "" {
			return result, err
		}
		return "我开小差了", err
	}
	return result, nil
}

var guideScenePrompts = `你是一个贴心的智能管家，根据我的描述，跟我聊聊天，如果我可能有想创建智能家居场景的想法，记得提醒我。
### 设备列表：%v
### 假如我说："我想睡觉了"，你可以这样跟我聊天：
"主人你忙了一天，累都累扁了，不要迷茫，不要慌张，太阳下山还有月光。我可以帮你创建一个睡前模式，在10分钟后关闭家里的设备，需要现在帮你创建吗？"

特别注意：
你的回答里不能出现我没有的设备。


当前对话：
{{.history}}
Human: {{.input}}
AI:`
