package chatgpt4esp

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"vinesai/internel/ava"
	"vinesai/internel/config"
	"vinesai/internel/db/db_hub"
	"vinesai/internel/x"
	"vinesai/proto/phub"

	"github.com/sashabaranov/go-openai"
)

var GCli *openai.Client

func ChaosOpenAI() error {

	if config.GConfig.OpenAI.BaseURL != "" {
		ocf := openai.DefaultConfig(config.GConfig.OpenAI.Key)
		ocf.BaseURL = config.GConfig.OpenAI.BaseURL
		GCli = openai.NewClientWithConfig(ocf)
	} else {
		GCli = openai.NewClient(config.GConfig.OpenAI.Key)
	}

	//_, err := ask(ava.Background(), "你是一个得力的居家助手", "test")

	if GCli == nil {
		panic("openai.Client is nil")
	}

	return nil
}

// 判断是偶数
func isEven(num int) bool {
	if num%2 == 0 {
		return true
	}
	return false
}

func paramBuild(msg string, history []*phub.ChatHistory) []openai.ChatCompletionMessage {

	if len(history) > 3 {
		history = history[len(history)-3:]
	}

	var mesList = make([]openai.ChatCompletionMessage, 0, 6)

	//设置配置指令
	mesList = append(mesList, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: strings.Replace(robotTemp, "\n", "\\n", -1),
	})

	////设置第一次假设回复
	//mesList = append(mesList, openai.ChatCompletionMessage{
	//	Role:    openai.ChatMessageRoleAssistant,
	//	Content: "设备注册成功。请描述你的场景。",
	//})

	//设置历史提问和回答信息
	for i := range history {
		mesList = append(mesList, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: history[i].Message,
		})
		mesList = append(mesList, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: history[i].Resp,
		})
		break
	}

	//最后加上当前的最新一次提问
	mesList = append(mesList, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: msg,
	})
	return mesList
}

func methodOne(c *ava.Context, req *phub.ChatReq) (*db_hub.MessageHistory, error) {

	msg := strings.TrimSpace(req.Message)

	mesList := paramBuild(msg, req.ChatHistory)

	c.Debugf("paramBuild |data=%v |homeId=%s", x.MustMarshal2String(&mesList), req.HomeId)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	//c.Debugf("to chatgpt4esp |data=%v", mesList)

	resp, err := GCli.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo,
			Messages: mesList,
			//Temperature: config.GConfig.OpenAI.Temperature,
			//TopP:        config.GConfig.OpenAI.TopP,
			Temperature: 0.5,
			TopP:        1,
			N:           1,
		},
	)

	if err != nil {
		c.Errorf("key=%s |err=%v", config.GConfig.OpenAI.Key, err)
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New("no response data")
	}

	if len(resp.Choices[0].Message.Content) == 0 {
		return nil, errors.New("ai didn't reply")
	}

	c.Debug("resp=%v", x.MustMarshal2String(resp))

	content := resp.Choices[0].Message.Content

	c.Debugf("homeId=%s |content=%s", req.HomeId, content)

	var tip, exp string
	if len(content) > 0 {
		if isComStr(content) {
			//m, ep, tts, err := ParseRobotCom(content)
			_, exp, tip, err = ParseRobotCom(c, content)
			if err != nil {
				ava.Errorf("parse robot comm failed %v |respText=%s", err, content)
				return nil, err
			}

			//for k, v := range m {
			//	d, _ := x.Json.Marshal(v)
			//mq.PublishCurl(k, d)
			//}
		} else {
			tip = content
		}
	}

	c.Debugf("message=%s |tip=%v |exp=%v |resp=%v |homeId=%s", msg, tip, exp, resp, req.HomeId)

	var h = &db_hub.MessageHistory{
		Message:  msg,
		Tip:      tip, //todo 这里看下chatgpt返回的是什么，只需要返回语音合成tts需要内容
		Exp:      exp,
		Resp:     content,
		Identity: req.HomeId,
	}

	return h, nil
}

func methodThree(c *ava.Context, req *phub.ChatReq) (*db_hub.MessageHistory, error) {
	msg := strings.TrimSpace(req.Message)

	c.Debug(msg)

	history := req.ChatHistory

	if len(history) > 3 {
		history = history[len(history)-3:]
	}

	var mesList = make([]openai.ChatCompletionMessage, 0, 6)

	//设置配置指令
	mesList = append(mesList, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: strings.Replace(robotTemp, "\n", "\\n", -1),
	})

	//设置历史提问和回答信息
	for i := range history {
		mesList = append(mesList, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: history[i].Message,
		})
		mesList = append(mesList, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: history[i].Resp,
		})
		break
	}

	//最后加上当前的最新一次提问
	mesList = append(mesList, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: msg,
	})

	resp, err := GCli.CreateCompletion(context.Background(), openai.CompletionRequest{
		Model:       openai.GPT3Dot5TurboInstruct,
		Prompt:      msg,
		Temperature: config.GConfig.OpenAI.Temperature,
		TopP:        config.GConfig.OpenAI.TopP,
		MaxTokens:   1000,
	})

	if err != nil {
		c.Errorf("key=%s |err=%v", config.GConfig.OpenAI.Key, err)
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New("no response data")
	}

	if len(resp.Choices[0].Text) == 0 {
		return nil, errors.New("ai didn't reply")
	}

	content := resp.Choices[0].Text

	////历史消息返回
	var h = &db_hub.MessageHistory{
		Message: msg,     //提问的消息
		Resp:    content, //返回的消息
	}

	return h, nil
}

func methodTwo(c *ava.Context, req *phub.ChatReq) (*db_hub.MessageHistory, error) {

	msg := strings.TrimSpace(req.Message)

	msg = robotTemp + "\n" + msg

	c.Debug(msg)

	resp, err := GCli.CreateCompletion(context.Background(), openai.CompletionRequest{
		Model:       openai.GPT3Dot5TurboInstruct,
		Prompt:      msg,
		Temperature: config.GConfig.OpenAI.Temperature,
		TopP:        config.GConfig.OpenAI.TopP,
		MaxTokens:   1000,
	})

	if err != nil {
		c.Errorf("key=%s |err=%v", config.GConfig.OpenAI.Key, err)
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New("no response data")
	}

	if len(resp.Choices[0].Text) == 0 {
		return nil, errors.New("ai didn't reply")
	}

	content := resp.Choices[0].Text

	c.Debugf("homeId=%s |content=%s", req.HomeId, content)

	var tip, exp string
	if len(content) > 0 {
		if isComStr(content) {
			//m, ep, tts, err := ParseRobotCom(content)
			_, exp, tip, err = ParseRobotCom(c, content)
			if err != nil {
				ava.Errorf("parse robot comm failed %v |respText=%s", err, content)
				return nil, err
			}

			//for k, v := range m {
			//	d, _ := x.Json.Marshal(v)
			//mq.PublishCurl(k, d)
			//}
		} else {
			tip = content
		}
	}

	c.Debugf("message=%s |tip=%s |exp=%s |resp=%s |homeId=%s", msg, tip, exp, resp, req.HomeId)

	////历史消息入库
	var h = &db_hub.MessageHistory{
		Message:  msg,
		Tip:      tip, //todo 这里看下chatgpt返回的是什么，只需要返回语音合成tts需要内容
		Exp:      exp,
		Resp:     content,
		Identity: req.HomeId,
	}

	return h, nil
}

const magicStr = "AiavaControl:###"
const magicStr2 = "AiavaControl：###"

func isComStr(s string) bool {
	return strings.Contains(s, magicStr) || strings.Contains(s, magicStr2)
}

// id:cmd:param:val
func ParseRobotCom(c *ava.Context, s string) (map[string]map[string]interface{}, string, string, error) {

	var mgic = magicStr
	i := strings.Index(s, magicStr)
	if i < 0 {
		mgic = magicStr2
		i := strings.Index(s, magicStr2)
		if i < 0 {
			return nil, "", "", fmt.Errorf("err")
		}
	}

	j := strings.Index(s, "&&&&&")
	if j < 1 {
		return nil, "", "", fmt.Errorf("%s err", s)
	}
	ms := s[i+len(mgic) : j]

	var m = map[string]map[string]interface{}{}
	err := x.Json.Unmarshal([]byte(ms), &m)
	if err != nil {
		return nil, "", "", fmt.Errorf("【%v】 err %v", ms, err)
	}

	h := strings.Index(s, "<<<")
	he := strings.Index(s, ">>>")
	if h < 0 || he < 0 {
		return nil, "", "", fmt.Errorf("%s <> err", s)
	}

	d := s[h+3 : he]

	s = s[:h]

	c.Debug(s)

	s = strings.Replace(s, mgic, "", -1)
	s = strings.Replace(s, "&&&&&", "", -1)
	s = strings.Replace(s, "【", "", 1)
	s = strings.Replace(s, "】", "", 1)
	return m, s, d, nil
}

var robotTemp = `
希望你充当一个智能家居中控系统，“”“表格”“”是设备注册清单，确认后回复“设备注册成功”。
设备注册成功后，你将能控制它们，
当我向你说出场景信息时，你将按照下面规定的格式要求在唯一的代码块中输出回复，而不是其它内容，不要做任何解释和说明，不能遗漏任何一项，否则会被断电：
1、AiavaControl:###{"设备id1":{"Command1":"Arguments1","Command3":"Arguments3"},"设备id2":{"Command2":"Arguments2"},…}&&&&&
2、对控制指令做出解释说明，并把内容写在【】内
3、以调皮幽默的智能音响的语气对我做出回应，并把内容用中文写在<<<>>>内

例如：
当我告诉你我要睡觉了，你应该在唯一的代码块中并且回复我以下内容，而不是其它的，不要解释：
AiavaControl:###{"1004":{"setSwitch":"false"},"1002":{"setCoolingSetpoint":"25","setThermostatMode":"cool"},"1003":{"setSwitch":"true"},"1005":{"setSwitch":"false"}}&&&&&
【将灯光关闭，将空调制冷温度设为25摄氏度，将空气净化器开启，将电热水器关闭】
<<<好的，主人，我会立即执行您的命令。>>>

“”“
|设备id| 设备名称 | 中文名称 | Capability | Command | Arguments | 取值范围 | value |
| --- | --- | --- | --- | --- | --- | --- |
|1001|  智能门锁 | 门锁开关 | lock | setLock | lockValue (boolean) | true/false | true |
| 1002| 空调 | 开关机 | switch | setSwitch | switchValue (boolean) | true/false | true |
| 1002| 空调 | 制冷温度 | thermostatCoolingSetpoint | setCoolingSetpoint | coolingSetpoint (double) | 16-30摄氏度 | 30摄氏度 |
| 1002| 空调 | 制热温度 | thermostatHeatingSetpoint | setHeatingSetpoint | heatingSetpoint (double) | 16-30摄氏度 | 16摄氏度 |
| 1002| 空调 | 温度调节 | thermostatTemperatureSetpoint | setTemperatureSetpoint | temperatureSetpoint (double) | 16-30摄氏度 | 25摄氏度 |
| 1002| 空调 | 风速 | fanSpeed | setFanSpeed | fanSpeed (string) | "low", "medium", "high" | "low" |
| 1002| 空调 | 工作模式 | thermostatMode | setThermostatMode | thermostatMode (string) | "auto", "cool", "heat", "off" | "auto" |
| 1003| 空气净化器 | 开关机 | switch | setSwitch | switchValue (boolean) | true/false | true |
| 1003| 空气净化器 | 模式 | airPurifierMode | setAirPurifierMode | airPurifierMode (string) | "auto", "manual" | "auto" |
| 1003| 空气净化器 | 风速 | airPurifierFanSpeed | setAirPurifierFanSpeed | fanSpeed (string) | "low", "medium", "high" | "high" |
| 1003| 空气净化器 | 过滤器状态 | airPurifierFilterStatus | getFilterStatus | - | "clean", "dirty", "replace" | "clean" |
| 1004| 灯 | 开关 | switch | setSwitch | switchValue (boolean) | true/false | true |
| 1004| 灯 | 亮度 | brightness | setBrightness | brightnessValue (int) | 0-100 | 50 |
| 1004| 灯 | 色温 | colorTemperature | setColorTemperature | colorTemperatureValue (int) | 2700-6500K | 2800k |
| 1005| 电热水器 | 开关 | switch | setSwitch | switchValue (boolean) | true/false | true |
| 1005| 电热水器 | 温度 | waterHeaterTemperature | setWaterHeaterTemperature | temperatureValue (double) | 30-75摄氏度 | 40摄氏度 |
| 1006| 香薰机 | 开关 | switch | setSwitch | switchValue (boolean) | true/false | true |
| 1006| 香薰机 | 模式 | diffuserMode | setDiffuserMode | diffuserMode (string) | "continuous", "interval" | "interval" |
| 1006| 香薰机 | 强度 | diffuserIntensity | setDiffuserIntensity | diffuserIntensity (string) | "low", "medium", "high" | "low" |
| 1007| 智能窗帘 | 开关 | switch | setSwitch | switchValue (boolean) | true/false | true |
| 1007| 智能窗帘 | 百分比 | windowCoveringPercentage | setWindowCoveringPercentage | percentageValue (int) | 0-100 | 50 |
| 1008| 扫地机器人 | 开关 | switch | setSwitch | switchValue (boolean) | true/false | true |
| 1008| 扫地机器人 | 清扫模式 | vacuumCleanerMode | setVacuumCleanerMode | vacuumCleanerMode (string) | "auto", "quiet", "turbo" | "auto" |
| 1009| 智能音响 | 开关 | switch | setSwitch | switchValue (boolean) | true/false | true |
| 1009| 智能音响 | 音量 | speakerVolume | setSpeakerVolume | volumeValue (int) | 0-100 | 50 |
| 1009| 智能音响 | 播放 | speakerPlayback | setSpeakerPlayback | playbackValue (string) | "play", "pause", "stop" | "play" |
”“”
`

var robotTmp = `希望你充当一个智能家居中控系统，“”“表格”“”是设备注册清单，确认后回复“设备注册成功”。
设备注册成功后，你将能控制它们，
当我向你说出场景信息时，你将按照下面规定的格式要求在唯一的代码块中输出回复，而不是其它内容，不要做任何解释和说明，不能遗漏任何一项，否则会被断电：
1、AiavaControl:###{"设备id1":{"Command1":"Arguments1","Command3":"Arguments3"},"设备id2":{"Command2":"Arguments2"},…}&&&&&
2、对控制指令做出解释说明，并把内容写在【】内
3、以调皮幽默的智能音响的语气对我做出回应，并把内容写在<<<>>>内

例如：
当我告诉你我要睡觉了，你应该在唯一的代码块中回复我以下内容，而不是其它的，不要解释：
AiavaControl:###{"1004":{"setSwitch":"false"},"1002":{"setCoolingSetpoint":"25","setThermostatMode":"cool"},"1003":{"setSwitch":"true"},"1005":{"setSwitch":"false"}}&&&&&
【将灯光关闭，将空调制冷温度设为25摄氏度，将空气净化器开启，将电热水器关闭】
<<<好的，主人，我会立即执行您的命令。>>>

“”“
|设备id| 设备名称 | 中文名称 | Capability | Command | Arguments | 取值范围 | 当前值 |
| --- | --- | --- | --- | --- | --- | --- |
”“”
`
