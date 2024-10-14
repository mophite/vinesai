package tuya

import (
	"context"
	"fmt"
	"time"
	"vinesai/internel/ava"
	"vinesai/internel/x"

	"github.com/sashabaranov/go-openai"
)

var botTmp = `根据我的意图描述，将其映射为具体的设备控制指令。

[设备清单]包含以下设备信息：
%s

[设备控制指令]要求如下：
%s

请根据以上信息生成相应的设备控制命令，并严格按照以下JSON格式返回，且JSON前后不要出现任何内容，否则视为非法数据;
如果无法根据用户描述推断出合理的控制指令，返回标准的错误响应;

**返回格式**（必须严格遵守以下格式，否则视为非法数据）：
- "voice" 字段：对应用户语音响应，用淘气的语气，包含对某个设备的控制详情
- "result" 字段：包含要控制的设备列表及具体指令
- "result_type" 字段：包含控制类型，0:表示设备控制，1:表示其他对话

### 正确的控制设备指令JSON 示例格式：
{
    "voice": "好的主人，客厅双色温筒灯已为你打开，并将亮度调到了30；客厅多彩射灯已为你打开，并将亮度调到了20",
	"result_type":0,
    "result": [
       {
         "id": "device_id_1",
         "name":"客厅双色温筒灯",
         "data": {
           "commands": [
             {"code": "switch_led", "value": true},
             {"code": "bright", "value": 30}
           ]
         }
       },
       {
         "id": "device_id_2",
         "name":"客厅多彩射灯",
         "data": {
           "commands": [
             {"code": "switch_led", "value": true},
             {"code": "bright", "value": 20}
           ]
         }
       }
    ]
}

### 正确的普通对话模式示例格式：
{
    "voice": "我最帅！",
	"result":[],
	"result_type":1
}

### 如果无法生成控制指令，请返回以下 JSON：
{
    "voice": "无法生成控制指令"
}

要求：
1. 输出的 JSON 必须严格按照示例格式，包括字段的大小写、逗号、引号和括号。
2. 开头和结尾不应包含任何多余的字符或空行。
3. 确保所有指令和设备控制逻辑准确无误。`

var filterDevice = `根据用户的意图，从家里的设备中筛选出需要控制的设备，并提取其设备 ID。

[设备清单]信息如下:
%v

根据我的意图进行筛选，并严格按照以下 JSON 格式输出。如果没有符合的设备，请输出空的 "devices" 数组。格式必须严格遵守以下示例格式，否则会被视为非法数据：

### JSON 输出示例：
{
	"voice":"",
    "devices": [
        {"id": "device_id_1", "name": "开关1"},
        {"id": "device_id_2", "name": "灯光2"}
    ]
}

### 如果没有符合条件的设备，输出以下格式：
{
	"voice":"没有复合条件的设备",
    "devices": []
}

要求：
1. 输出的 JSON 数据格式必须严格遵循示例，包括字段的大小写、标点符号等。
2. 输出中不得包含多余字符、空行或额外信息。
3. 确保设备的筛选逻辑准确无误，ID 和名称正确对应。`

var gOpenAi *openai.Client

var defaultKey = "sk-08cdfea5547040209ea0e2d874fff912"

//var defaultKey="sk-2RET3Pqa6Z3g6b0pE29351119e9b410fAfC3D44b4eC4C4A9"

var defaultUrl = "https://dashscope.aliyuncs.com/compatible-mode/v1"

//var defaultUrl = "https://ai-yyds.com/v1"

func init() {
	ocf := openai.DefaultConfig(defaultKey)
	ocf.BaseURL = defaultUrl
	gOpenAi = openai.NewClientWithConfig(ocf)
}

type devicesFromGpt struct {
	Devices []*shortDevice `json:"devices"`
}

// 通过ai获取到需要控制的设备列表
func deviceListGpt(c *ava.Context, content string, devices []*device) ([]*shortDevice, map[string]*device, []string, error) {
	var msgList = make([]openai.ChatCompletionMessage, 0, 6)

	//设置配置指令
	msgList = append(msgList, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: "你是一个智能管家",
	})

	msgList = append(msgList, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: fmt.Sprintf(filterDevice, x.MustMarshal2String(devices)),
	})

	msgList = append(msgList, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: content,
	})

	c.Debugf("----------------------------deviceListGpt----------------,data=%s", fmt.Sprintf(filterDevice, x.MustMarshal2String(devices)))

	var now = time.Now()
	resp, err := gOpenAi.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			//Model:          "gpt-4o",
			//Model:          "gpt-4o-mini-2024-07-18",
			Model:          "qwen-turbo-latest",
			Messages:       msgList,
			Temperature:    0.1,
			TopP:           0.1,
			N:              1,
			ResponseFormat: &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject},
		},
	)

	c.Debugf("------------------通过ai获取到需要控制的设备列表-------------latecy=%v", time.Now().Sub(now).Seconds())

	if err != nil {
		c.Error(err)
		return nil, nil, nil, err
	}

	c.Debugf("FROM |data=%s", ava.MustMarshalString(&resp))

	str := resp.Choices[0].Message.Content

	var r devicesFromGpt

	ava.MustUnmarshal(ava.StringToBytes(str), &r)

	c.Debugf("resp=%s |result=%v", str, ava.MustMarshalString(r))

	l, ids := shortDeviceInfo2Devices(r.Devices, devices)
	return r.Devices, l, ids, nil
}

// 发送消息给ai
func msg2Gpt(c *ava.Context, content, commands, deviceList string) (*aiResp, error) {
	var msgList = make([]openai.ChatCompletionMessage, 0, 6)

	//设置配置指令
	msgList = append(msgList, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: "你是一个智能管家",
	})

	msgList = append(msgList, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: fmt.Sprintf(botTmp, deviceList, commands),
	})

	msgList = append(msgList, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: content,
	})

	var now = time.Now()
	resp, err := gOpenAi.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			//Model:          "gpt-4o",
			//Model:          "gpt-4o-mini-2024-07-18",
			Model:          "qwen-turbo-latest",
			Messages:       msgList,
			Temperature:    0.1,
			TopP:           0.1,
			N:              1,
			ResponseFormat: &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject},
		},
	)

	c.Debugf("------------------发送消息给ai-------------%v data=%s", time.Now().Sub(now).Seconds(), fmt.Sprintf(botTmp, deviceList, commands))

	if err != nil {
		c.Error(err)
		return nil, err
	}

	c.Debugf("FROM |data=%s", ava.MustMarshalString(&resp))

	str := resp.Choices[0].Message.Content

	var r = &aiResp{}

	ava.MustUnmarshal(ava.StringToBytes(str), &r)

	c.Debugf("resp=%s |result=%v", str, ava.MustMarshalString(r))

	return r, nil
}
