package homeassistant

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
	"vinesai/internel/ava"
	"vinesai/internel/x"
	"vinesai/proto/pha"

	"github.com/sashabaranov/go-openai"
	"vinesai/app/api/api.home/miniprogram"
)

type HomeAssistant struct {
}

func (h *HomeAssistant) CallWs(c *ava.Context, req *pha.CallReq, rsp *pha.CallRsp) {
	//if req.Home == "" {
	//	rsp.Code = http.StatusBadRequest
	//	rsp.Msg = "我不知道你要控制哪个家庭设备"
	//	return
	//}

	if req.Message == "" {
		rsp.Code = http.StatusBadRequest
		rsp.Msg = "我听不见"
		return
	}

	//获取指令集
	services, err := getServices(c, "123")
	if err != nil {
		c.Error(err)
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = "没有指令数据"
		return
	}

	states, _, err := getStates(c, "123")
	if err != nil {
		c.Error(err)
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = "没有数据"
		return
	}

	//c.Debugf("-----services-----%s", services)
	//c.Debugf("-----states-----%s", states)
	//告诉ai指令
	var top = openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: fmt.Sprintf(aiTmp2Ws, services, states),
	}

	var next = openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: req.Message,
	}

	var msgList = make([]openai.ChatCompletionMessage, 0, 3)
	msgList = append(msgList, top, next)

	resp, err := miniprogram.OpenAi.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			//Model:    openai.GPT3Dot5Turbo,
			//Model: "claude-3-haiku-20240307",
			Model:    "gpt-4o",
			Messages: msgList,
			//Temperature: config.GConfig.OpenAI.Temperature,
			//TopP:        config.GConfig.OpenAI.TopP,
			Temperature:    0.5,
			TopP:           1,
			N:              1,
			MaxTokens:      4000,
			ResponseFormat: &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject},
		},
	)

	if err != nil {
		c.Error(err)
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = "我有点晕。"
		return
	}

	c.Debugf("FROM |data=%s", ava.MustMarshalString(&resp))

	if len(resp.Choices) == 0 {
		c.Error(err)
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = "我犯迷糊了。"
		return
	}

	var result struct {
		Command []struct {
			ID          int64       `json:"id"`
			Type        string      `json:"type"`
			Domain      string      `json:"domain"`
			Service     string      `json:"service"`
			ServiceData interface{} `json:"service_data"`
			Target      interface{} `json:"target"`
		} `json:"serviceData"`
		Message string `json:"message"`
	}

	str := resp.Choices[0].Message.Content
	str = strings.ReplaceAll(str, "\n", "")
	str = strings.ReplaceAll(str, "\t", "")
	str = strings.ReplaceAll(str, `\`, "")

	ava.MustUnmarshal(ava.StringToBytes(str), &result)

	c.Debugf("text=%s |result=%v", str, ava.MustMarshalString(result))

	id := time.Now().UnixNano()

	for i := range result.Command {
		command := result.Command[i]
		command.ID = atomic.AddInt64(&id, 1)
		//发起设备控制
		callServiceWs("123", command)
	}

	rsp.Code = http.StatusOK
	rsp.Msg = result.Message
}

// Call todo 用户输入，控制设备
func (h *HomeAssistant) Call(c *ava.Context, req *pha.CallReq, rsp *pha.CallRsp) {
	//if req.Home == "" {
	//	rsp.Code = http.StatusBadRequest
	//	rsp.Msg = "我不知道你要控制哪个家庭设备"
	//	return
	//}

	if req.Message == "" {
		rsp.Code = http.StatusBadRequest
		rsp.Msg = "我听不见"
		return
	}

	//获取指令集
	services, err := getServices(c, "123")
	if err != nil {
		c.Error(err)
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = "没有指令数据"
		return
	}

	states, _, err := getStates(c, "123")
	if err != nil {
		c.Error(err)
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = "没有数据"
		return
	}

	//c.Debugf("-----services-----%s", services)
	//c.Debugf("-----states-----%s", states)
	//告诉ai指令
	var top = openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: fmt.Sprintf(aiTmp2, services, states),
	}

	var next = openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: req.Message,
	}

	var msgList = make([]openai.ChatCompletionMessage, 0, 3)
	msgList = append(msgList, top, next)

	resp, err := miniprogram.OpenAi.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			//Model:    openai.GPT3Dot5Turbo,
			//Model: "claude-3-haiku-20240307",
			Model:    "gpt-4o",
			Messages: msgList,
			//Temperature: config.GConfig.OpenAI.Temperature,
			//TopP:        config.GConfig.OpenAI.TopP,
			Temperature:    0.5,
			TopP:           1,
			N:              1,
			MaxTokens:      4000,
			ResponseFormat: &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject},
		},
	)

	if err != nil {
		c.Error(err)
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = "我有点晕。"
		return
	}

	c.Debugf("FROM |data=%s", ava.MustMarshalString(&resp))

	if len(resp.Choices) == 0 {
		c.Error(err)
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = "我犯迷糊了。"
		return
	}

	var result struct {
		Command []struct {
			Data    interface{} `json:"data"`
			Domain  string      `json:"domain"`
			Service string      `json:"service"`
		} `json:"serviceData"`
		Message string `json:"message"`
	}

	str := resp.Choices[0].Message.Content
	str = strings.ReplaceAll(str, "\n", "")
	str = strings.ReplaceAll(str, "\t", "")
	str = strings.ReplaceAll(str, `\`, "")

	ava.MustUnmarshal(ava.StringToBytes(str), &result)

	c.Debugf("text=%s |result=%v", str, ava.MustMarshalString(result))

	for i := range result.Command {
		command := result.Command[i]
		//发起设备控制
		callServiceHttp("123", command.Domain+"/"+command.Service, x.MustMarshal(command.Data))
	}

	rsp.Code = http.StatusOK
	rsp.Msg = result.Message
}

// 获取指令列表
func (h *HomeAssistant) Services(c *ava.Context, req *pha.ServicesReq, rsp *pha.ServicesRsp) {
	//home_id := c.GetHeader("home_id")
	//if home_id == "" || mapHome2Url[home_id] == "" {
	//	rsp.Code = http.StatusBadRequest
	//	rsp.Msg = "header中的home_id不能为空或当前home不存在"
	//	return
	//}
	data, err := getServices(c, "123")
	if err != nil {
		c.Error(err)
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = "没有数据"
		return
	}

	rsp.Data = &pha.ServicesData{
		Services: data,
	}

	rsp.Code = http.StatusOK

}

// 获取设备当前状态
func (h *HomeAssistant) States(c *ava.Context, req *pha.StatesReq, rsp *pha.StatesRsp) {
	//home_id := c.GetHeader("home_id")
	//if home_id == "" || mapHome2Url[home_id] == "" {
	//	rsp.Code = http.StatusBadRequest
	//	rsp.Msg = "header中的home_id不能为空或当前home不存在"
	//	return
	//}

	data, _, err := getStates(c, "123")
	if err != nil {
		c.Error(err)
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = "没有数据"
		return
	}

	rsp.Data = &pha.StatesData{
		States: data,
	}

	rsp.Code = http.StatusOK
}
