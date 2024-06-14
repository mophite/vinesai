package homeassistant

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"vinesai/internel/ava"
	"vinesai/internel/x"

	"github.com/sashabaranov/go-openai"
	"vinesai/app/api/api.home/miniprogram"
)

type command struct {
	EntityId string      `json:"entity_id"`
	Service  string      `json:"service"`
	Data     interface{} `json:"data"`
}

type Commands struct {
	Commands []command `json:"commands"`
	Answer   string    `json:"answer"`
}

type wsCallService struct {
	Id          int64       `json:"id"`
	Type        string      `json:"type"`
	Domain      string      `json:"domain"`
	Service     string      `json:"service"`
	ServiceData interface{} `json:"service_data"`
	Target      struct {
		EntityId string `json:"entity_id"`
	} `json:"target"`

	ReturnResponse bool `json:"return_response"`
}

type difyRsp struct {
	Data struct {
		CreatedAt   int64       `json:"created_at"`
		ElapsedTime float64     `json:"elapsed_time"`
		Error       interface{} `json:"error"`
		FinishedAt  int64       `json:"finished_at"`
		ID          string      `json:"id"`
		Outputs     struct {
			Text string `json:"text"`
		} `json:"outputs"`
		Status      string `json:"status"`
		TotalSteps  int64  `json:"total_steps"`
		TotalTokens int64  `json:"total_tokens"`
		WorkflowID  string `json:"workflow_id"`
	} `json:"data"`
	TaskID        string `json:"task_id"`
	WorkflowRunID string `json:"workflow_run_id"`
}

type difyReq struct {
	Inputs struct {
		Query string `json:"query"`
	} `json:"inputs"`
	ResponseMode string `json:"response_mode"`
	User         string `json:"user"`
}

const (
	difyUrl    = "http://223.72.19.182:8000/v1/workflows/run"
	difyBearer = "Bearer app-gmhMQjOwBtsyYMfvlcQDx0E0"
)

func SelectAI(c *ava.Context, home, message string) (*Commands, error) {
	//return chatgpt(c, home, message)
	return dify(c, home, message)
}

func dify(c *ava.Context, home, message string) (*Commands, error) {

	var difyReq difyReq
	difyReq.User = "abc-123"
	difyReq.ResponseMode = "blocking"
	difyReq.Inputs.Query = message

	req, _ := http.NewRequest("POST", difyUrl, bytes.NewReader(x.MustMarshal(&difyReq)))
	req.Header.Set("Authorization", difyBearer)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	c.Debugf("FROM |data=%s", x.BytesToString(body))

	var difyRsp difyRsp
	err = x.MustNativeUnmarshal(body, &difyRsp)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	c.Debugf("TO=%s |FROM=%s", x.MustMarshal2String(&difyReq), ava.MustMarshalString(&difyRsp))

	var result Commands

	str := difyRsp.Data.Outputs.Text
	str = strings.ReplaceAll(str, "\n", "")
	str = strings.ReplaceAll(str, "\t", "")
	str = strings.ReplaceAll(str, `\`, "")

	ava.MustUnmarshal(ava.StringToBytes(str), &result)

	c.Debugf("text=%s |result=%v", str, ava.MustMarshalString(&result))

	return &result, nil

}

func chatgpt(c *ava.Context, home, message string) (*Commands, error) {
	//获取指令集
	services, err := getServices(c, home)
	if err != nil {
		c.Error(err)
		return nil, errors.New("没有指令数据")
	}

	states, _, err := getStates(c, home)
	if err != nil {
		c.Error(err)
		return nil, errors.New("没有设备数据")
	}

	//告诉ai指令
	var top = openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: fmt.Sprintf(aiTmp, services, states),
	}

	var next = openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message,
	}

	var msgList = make([]openai.ChatCompletionMessage, 0, 3)
	msgList = append(msgList, top, next)

	resp, err := miniprogram.OpenAi.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			//Model: openai.GPT3Dot5Turbo,
			//Model: "claude-3-haiku-20240307",
			Model:    "gpt-4o",
			Messages: msgList,
			//Temperature: config.GConfig.OpenAI.Temperature,
			//TopP:        config.GConfig.OpenAI.TopP,
			Temperature: 0.5,
			TopP:        1,
			N:           1,
			//MaxTokens:      4000,
			ResponseFormat: &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject},
		},
	)

	if err != nil {
		c.Error(err)
		return nil, err
	}

	c.Debugf("FROM |data=%s", ava.MustMarshalString(&resp))

	if len(resp.Choices) == 0 {
		c.Error(err)
		return nil, errors.New("返回数据为空")
	}

	var result Commands

	str := resp.Choices[0].Message.Content
	str = strings.ReplaceAll(str, "\n", "")
	str = strings.ReplaceAll(str, "\t", "")
	str = strings.ReplaceAll(str, `\`, "")

	ava.MustUnmarshal(ava.StringToBytes(str), &result)

	c.Debugf("text=%s |result=%v", str, ava.MustMarshalString(result))

	return &result, nil
}
