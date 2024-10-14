package langchain

import "context"

type chat struct {
}

func (c *chat) Name() string {
	return "chatAgent"
}

func (c *chat) Description() string {
	return `日常对话，无其他特殊功能`
}

func (c *chat) Call(ctx context.Context, input string) (string, error) {

	return "天气很好", nil
}
