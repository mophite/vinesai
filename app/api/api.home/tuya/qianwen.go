package tuya

import (
	"fmt"
	"vinesai/internel/ava"
	"vinesai/internel/lib"
	"vinesai/internel/x"
)

// 阿里千问
var defaultQianWenKey = "sk-08cdfea5547040209ea0e2d874fff912"
var defaultQianwenUrl = "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
var qianwenHeader = map[string]string{
	"Authorization": "Bearer sk-08cdfea5547040209ea0e2d874fff912",
	"Content-Type":  "application/json",
}

type qianwenResp struct {
	Choices []struct {
		FinishReason string      `json:"finish_reason"`
		Index        int64       `json:"index"`
		Logprobs     interface{} `json:"logprobs"`
		Message      struct {
			Content string `json:"content"`
			Role    string `json:"role"`
		} `json:"message"`
	} `json:"choices"`
	Created           int64       `json:"created"`
	ID                string      `json:"id"`
	Model             string      `json:"model"`
	Object            string      `json:"object"`
	SystemFingerprint interface{} `json:"system_fingerprint"`
	Usage             struct {
		CompletionTokens int64 `json:"completion_tokens"`
		PromptTokens     int64 `json:"prompt_tokens"`
		TotalTokens      int64 `json:"total_tokens"`
	} `json:"usage"`
}

// 通过ai获取到需要控制的设备列表
func deviceListQianwen(c *ava.Context, content string, devices []*device) ([]*shortDevice, map[string]*device, []string, error) {
	requestBody := map[string]interface{}{
		"temperature": 0.1,
		"top_p":       0.1,
		"model":       "qwen-turbo-latest",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "你是一个智能管家",
			},
			{
				"role":    "system",
				"content": fmt.Sprintf(filterDevice, x.MustMarshal2String(devices)),
			},
			{
				"role":    "user",
				"content": content,
			},
		},
	}

	c.Debugf("-----------------千问测试----------------%s", x.MustMarshal2String(requestBody))

	data, err := lib.POST(c, defaultQianwenUrl, x.MustMarshal(requestBody), qianwenHeader)
	if err != nil {
		c.Error(err)
		return nil, nil, nil, err
	}

	var resp qianwenResp
	err = x.MustUnmarshal(data, &resp)
	if err != nil {
		c.Error(err)
		return nil, nil, nil, err
	}

	c.Debugf("FROM |data=%s", string(data))

	str := resp.Choices[0].Message.Content

	var r devicesFromGpt

	ava.MustUnmarshal(ava.StringToBytes(str), &r)

	c.Debugf("resp=%s |result=%v", str, ava.MustMarshalString(r))

	l, ids := shortDeviceInfo2Devices(r.Devices, devices)
	return r.Devices, l, ids, nil
}

// 发送消息给ai
func msg2Qianwen(c *ava.Context, content, commands, deviceList string) (*aiResp, error) {
	requestBody := map[string]interface{}{
		"temperature": 0.1,
		"top_p":       0.1,
		"model":       "qwen-turbo-latest",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "你是一个智能管家",
			},
			{
				"role":    "system",
				"content": fmt.Sprintf(botTmp, deviceList, commands),
			},
			{
				"role":    "user",
				"content": content,
			},
		},
	}

	c.Debugf("-----------------千问测试----------------%s", x.MustMarshal2String(requestBody))

	data, err := lib.POST(c, defaultQianwenUrl, x.MustMarshal(requestBody), qianwenHeader)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	var resp qianwenResp
	err = x.MustUnmarshal(data, &resp)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	c.Debugf("FROM |data=%s", string(data))

	str := resp.Choices[0].Message.Content

	var r = &aiResp{}

	ava.MustUnmarshal(ava.StringToBytes(str), &r)

	c.Debugf("resp=%s |result=%v", str, ava.MustMarshalString(r))

	return r, nil
}
