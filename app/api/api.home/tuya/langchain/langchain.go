package langchain

import (
	"context"
	"regexp"
	"sync"
	"vinesai/internel/ava"
	"vinesai/internel/lib/tuyago"
	"vinesai/internel/x"

	"vinesai/internel/langchaingo/agents"
	"vinesai/internel/langchaingo/chains"
	"vinesai/internel/langchaingo/llms"
	"vinesai/internel/langchaingo/llms/openai"
	"vinesai/internel/langchaingo/memory"
	"vinesai/internel/langchaingo/schema"
	"vinesai/internel/langchaingo/tools"

	"github.com/pkg/errors"
)

var Tools = []tools.Tool{
	&syncDevices{},
	&summary{CallbacksHandler: LogHandler{}},
	&queryOnline{CallbacksHandler: LogHandler{}},
	&queryOffline{CallbacksHandler: LogHandler{}},
	&queryDevice{CallbacksHandler: LogHandler{}},
	&scene{CallbacksHandler: LogHandler{}},
	&guide{CallbacksHandler: LogHandler{}},
	&sceneQuery{CallbacksHandler: LogHandler{}},
	&autoQuery{CallbacksHandler: LogHandler{}},
	&runScene{CallbacksHandler: LogHandler{}},
}

var defaultKey = "sk-08cdfea5547040209ea0e2d874fff912"
var defaultUrl = "https://dashscope.aliyuncs.com/compatible-mode/v1"
var modelAgent = "qwen-turbo"
var modeGuide = "qwen-plus"
var modeLongText = "qwen-long"

//
//var defaultKey = "sk-2RET3Pqa6Z3g6b0pE29351119e9b410fAfC3D44b4eC4C4A9"
//var defaultUrl = "https://ai-yyds.com/v1"
//var modelAgent = "claude-3-5-sonnet-20241022"
//var modeGuide = "gpt-4o"
//var modeLongText = "claude-3-5-sonnet-20241022"

var langchaingoOpenAi *openai.LLM
var langchaingoGuideOpenAi *openai.LLM
var llmOpenAi *openai.LLM
var newExecutor *agents.Executor

func init() {
	var err error
	langchaingoOpenAi, err = openai.New(
		openai.WithBaseURL(defaultUrl),
		openai.WithToken(defaultKey),
		openai.WithModel(modelAgent),
		openai.WithCallback(LogHandler{}),
	)

	langchaingoGuideOpenAi, err = openai.New(
		openai.WithBaseURL(defaultUrl),
		openai.WithToken(defaultKey),
		openai.WithModel(modeGuide),
		openai.WithCallback(LogHandler{}),
	)

	llmOpenAi, err = openai.New(
		openai.WithBaseURL(defaultUrl),
		openai.WithToken(defaultKey),
		openai.WithModel(modeLongText),
		openai.WithResponseFormat(openai.ResponseFormatJSON),
		openai.WithCallback(LogHandler{}),
	)

	if err != nil {
		panic(err)
	}

	tuyago.Register("deviceOffline", &deviceOffline{})                 //设备离线
	tuyago.Register("deviceOnline", &deviceOnline{})                   //设备上线
	tuyago.Register("deviceBindSpace", &deviceBindSpace{})             //设备绑定
	tuyago.Register("devicePropertyMessage", &devicePropertyMessage{}) //设备状态上报
	tuyago.Register("deviceUnbindSpace", &deviceUnbindSpace{})         //设备解绑
}

func findJSON(str string) string {
	// 正则表达式用于匹配大括号内的文本，可能包含空格和换行符
	re := regexp.MustCompile(`(?s)\{.*\}`)
	matches := re.FindAllString(str, -1)
	if len(matches) == 0 {
		return ""
	}
	return matches[0]
}

func GenerateContentTurbo(c *ava.Context, prompt, input string) (string, error) {

	conversation := chains.NewConversation(
		langchaingoGuideOpenAi,
		memory.NewConversationBuffer(memory.WithChatHistory(buffChatMemory)), prompt,
	)
	result, err := chains.Run(
		context.Background(),
		conversation,
		input,
		chains.WithCallback(LogHandler{}),
		chains.WithTopP(0.5),
		chains.WithTemperature(0.5),
	)

	if err != nil {
		c.Error(err)
		return "服务器开小差了", err
	}

	if len(result) == 0 {
		return "我没有什么话说", errors.New("ai no resp")
	}

	c.Debugf("input=%s |resutl=%s", input, result)

	return result, nil
}

func GenerateContentWithout(c *ava.Context, mcList []llms.MessageContent, v interface{}) error {
	resp, err := llmOpenAi.GenerateContent(
		context.Background(),
		mcList,
		llms.WithTemperature(0.5),
		llms.WithN(1),
		llms.WithTopP(0.5),
	)

	if err != nil {
		c.Error(err)
		return err
	}

	if len(resp.Choices) == 0 {
		return errors.New("ai no resp")
	}

	content := findJSON(resp.Choices[0].Content)

	err = x.MustNativeUnmarshal([]byte(content), v)
	if err != nil {
		c.Error(err)
		return err
	}

	//c.Debugf("ai resp=%s |data=%s |content=%s", content, x.MustMarshal2String(v), resp.Choices[0].Content)

	return nil
}

var buffChatMemory = newBufferChatMessageHistory(withBufferChatMessageHistoryLimit(5))

func fromCtx(ctx context.Context) *ava.Context {
	c, _ := ctx.Value(defaultAvaCtxKey).(*ava.Context)
	return c
}

var defaultFirstInputKey = "x-langchain-first-input"

func getFirstInput(c *ava.Context) string {
	return c.GetString(defaultFirstInputKey)
}

func setFirstInput(c *ava.Context, input string) {
	c.Set(defaultFirstInputKey, input)
}

func getHomeId(c *ava.Context) string {
	return c.GetString(defaultHomeIdKey)
}

func setHomeId(c *ava.Context, input string) {
	c.Set(defaultHomeIdKey, input)
}

// 附带用户设备的品类
func Run(c *ava.Context, uid, homeId, content string) (string, error) {
	setHomeId(c, homeId)
	setFirstInput(c, content)

	ctx := context.WithValue(context.Background(), defaultBufferUidKey, uid)
	ctx = context.WithValue(ctx, defaultAvaCtxKey, c)

	var prompt = runSystemPromptsWithHistory
	if buffChatMemory.Len(ctx) == 0 {
		prompt = runSystemPrompts
	}

	a := agents.NewOpenAIFunctionsAgent(langchaingoOpenAi,
		Tools,
		agents.NewOpenAIOption().WithSystemMessage(prompt),
	)

	newExecutor = agents.NewExecutor(
		a,
		agents.WithCallbacksHandler(LogHandler{}),
		agents.WithMemory(memory.NewConversationBuffer(memory.WithChatHistory(buffChatMemory))),
		agents.WithMaxIterations(1),
	)

	result, err := chains.Run(
		ctx,
		newExecutor,
		content,
		chains.WithCallback(LogHandler{}),
		chains.WithTopP(0.5),
		chains.WithTemperature(0.5),
	)

	if err != nil && !errors.Is(err, agents.ErrNotFinished) {
		c.Errorf("result=%v |err=%v", result, err)
		if result != "" {
			return result, err
		}
		return "出了点小故障，请重试", err
	}

	return result, err
}

type Response struct {
	Voice string `json:"voice"`
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
var defaultAvaCtxKey = "x-langchaingo-ctx"
var defaultHomeIdKey = "x-langchaingo-homeid"

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
	ava.Debugf("buffChatMessage |AddUserMessage |data=%v", message)

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

func (c *buffChatMessage) Len(ctx context.Context) int {
	return len(c.Message[c.getUID(ctx)])
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

	ava.Debugf("buffChatMessage |message |data=%v", x.MustMarshal2String(messages))

	return nil
}

var runSystemPromptsWithHistory = `你名字叫小冰，是一个性格俏皮的人性化管家助理；担当智能家居设备控制和其他生活管理、咨询的工作。
如果没有找到我给你的能力，就当作普通对话不要使用Function Calling能力。
上一次对话记录：
{{.history}}`

var runSystemPrompts = `你名字叫小冰，是一个性格俏皮的人性化管家助理，担当智能家居设备控制和其他生活管理、咨询的工作。
如果没有找到我给你的能力，就当作普通对话不要使用Function Calling能力。`
