package tuya

import (
	"context"
	"fmt"
	"sync"
	"vinesai/internel/ava"
	"vinesai/internel/x"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/schema"
)

var langchaingoOpenAi *openai.LLM

func init() {
	var err error
	langchaingoOpenAi, err = openai.New(
		openai.WithBaseURL(defaultUrl),
		openai.WithToken(defaultKey),
		openai.WithModel("qwen-turbo-latest"),
		openai.WithResponseFormat(openai.ResponseFormatJSON),
	)
	if err != nil {
		panic(err)
	}

}

var buffChatMemory = newBufferChatMessageHistory()

func langchainRun(c *ava.Context, uid, content, deviceList, commands string) (*aiResp, error) {

	msgList := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "你是一个智能家居管家，具备有趣的灵魂"),
		llms.TextParts(llms.ChatMessageTypeHuman, fmt.Sprintf(botTmp, deviceList, commands)),
		llms.TextParts(llms.ChatMessageTypeHuman, content),
	}

	chatHistory := memory.NewConversationBuffer(
		memory.WithChatHistory(buffChatMemory),
	)

	llmChain := chains.NewConversation(langchaingoOpenAi, chatHistory)
	ctx := context.WithValue(context.Background(), defaultBufferUidKey, uid)

	out, err := chains.Run(
		ctx,
		llmChain,
		x.MustMarshal2String(msgList),
		chains.WithTopP(0.1),
		chains.WithTemperature(0.1),
	)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	var resp aiResp
	err = x.MustNativeUnmarshal([]byte(out), &resp)
	if err != nil {
		c.Error(err)
		return nil, nil
	}

	return &resp, nil
}

type buffChatMessage struct {
	Limit   int
	Message map[string][]llms.ChatMessage //一个用户一个记录
	lock    *sync.RWMutex
}

type bufferChatOption func(m *buffChatMessage)

func newBufferChatMessageHistory(options ...bufferChatOption) *buffChatMessage {
	return applyChatOptions(options...)
}

func applyChatOptions(options ...bufferChatOption) *buffChatMessage {
	h := &buffChatMessage{
		Message: make(map[string][]llms.ChatMessage, 100),
		lock:    new(sync.RWMutex),
	}

	for _, option := range options {
		option(h)
	}

	if h.Limit < 1 {
		h.Limit = 3
	}

	return h
}

func withBufferChatMessageHistoryLimit(h int) bufferChatOption {
	return func(m *buffChatMessage) {
		m.Limit = h
	}
}

var _ schema.ChatMessageHistory = &buffChatMessage{}

var defaultBufferUidKey = "x-langchaingo-uid"

func (c *buffChatMessage) setUID(ctx context.Context, uid string) {
	ctx = context.WithValue(ctx, defaultBufferUidKey, uid)
}

func (c *buffChatMessage) getUID(ctx context.Context) string {
	v := ctx.Value(defaultBufferUidKey)
	if v == nil {
		return ""
	}
	return v.(string)
}

func (c *buffChatMessage) AddMessage(ctx context.Context, message llms.ChatMessage) error {
	ava.Debugf("buffChatMessage |AddMessage |data=%v", message)
	c.lock.Lock()
	if len(c.Message[c.getUID(ctx)]) >= c.Limit {
		copy(c.Message[c.getUID(ctx)][:len(c.Message[c.getUID(ctx)])-1], c.Message[c.getUID(ctx)][1:])
		c.Message[c.getUID(ctx)][len(c.Message[c.getUID(ctx)])-1] = message // 在切片末尾添加新元素
	} else {
		c.Message[c.getUID(ctx)] = append(c.Message[c.getUID(ctx)], message)
	}
	c.lock.Unlock()
	return nil
}

func (c *buffChatMessage) AddUserMessage(ctx context.Context, message string) error {
	ava.Debugf("buffChatMessage |AddAIMessage |data=%v", message)

	c.lock.Lock()
	if len(c.Message[c.getUID(ctx)]) >= c.Limit {
		copy(c.Message[c.getUID(ctx)][:len(c.Message[c.getUID(ctx)])-1], c.Message[c.getUID(ctx)][1:])
		c.Message[c.getUID(ctx)][len(c.Message[c.getUID(ctx)])-1] = llms.HumanChatMessage{Content: message} // 在切片末尾添加新元素
	} else {
		c.Message[c.getUID(ctx)] = append(c.Message[c.getUID(ctx)], llms.HumanChatMessage{Content: message})
	}
	c.lock.Unlock()
	return nil
}

func (c *buffChatMessage) AddAIMessage(ctx context.Context, message string) error {

	c.lock.Lock()
	if len(c.Message[c.getUID(ctx)]) >= c.Limit {
		copy(c.Message[c.getUID(ctx)][:len(c.Message[c.getUID(ctx)])-1], c.Message[c.getUID(ctx)][1:])
		c.Message[c.getUID(ctx)][len(c.Message[c.getUID(ctx)])-1] = llms.AIChatMessage{Content: message} // 在切片末尾添加新元素
	} else {
		c.Message[c.getUID(ctx)] = append(c.Message[c.getUID(ctx)], llms.AIChatMessage{Content: message})
	}
	c.lock.Unlock()

	ava.Debugf("buffChatMessage |AddAIMessage |data=%v", message)

	return nil
}

func (c *buffChatMessage) Clear(ctx context.Context) error {
	c.lock.Lock()
	c.Message[c.getUID(ctx)] = nil
	c.Message[c.getUID(ctx)] = make([]llms.ChatMessage, c.Limit)
	c.lock.Unlock()
	return nil
}

func (c *buffChatMessage) Messages(ctx context.Context) ([]llms.ChatMessage, error) {
	c.lock.RLock()
	var r []llms.ChatMessage
	r = c.Message[c.getUID(ctx)]
	c.lock.RUnlock()

	ava.Debugf("buffChatMessage |message |data=%v", x.MustMarshal2String(r))

	return r, nil
}

func (c *buffChatMessage) SetMessages(ctx context.Context, messages []llms.ChatMessage) error {
	_ = c.Clear(ctx)

	c.lock.Lock()
	c.Message[c.getUID(ctx)] = append(c.Message[c.getUID(ctx)], messages...)
	if len(c.Message[c.getUID(ctx)]) > 3 {
		c.Message[c.getUID(ctx)] = c.Message[c.getUID(ctx)][:3]
	}
	c.lock.Unlock()
	return nil
}
