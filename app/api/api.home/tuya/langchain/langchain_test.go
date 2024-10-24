package langchain

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"vinesai/internel/ava"
	"vinesai/internel/config"
	"vinesai/internel/lib/tuyago"

	"vinesai/internel/langchaingo/embeddings"
	"vinesai/internel/langchaingo/llms/openai"
	"vinesai/internel/langchaingo/prompts"
	"vinesai/internel/langchaingo/schema"
	"vinesai/internel/langchaingo/vectorstores"
	"vinesai/internel/langchaingo/vectorstores/redisvector"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func init() {
	ava.SetupService(
		ava.Namespace("api.home"),
		ava.HttpApiAdd("0.0.0.0:10010"),
		//ava.TCPApiPort(10001),
		//ava.WssApiAddr("0.0.0.0:10002", "/ws"),
		ava.EtcdConfig(&clientv3.Config{Endpoints: []string{"47.106.129.170:2379"}}),
		//ava.WatchDog(tuya.Authorization),
		ava.ConfigOption(
			ava.Chaos(
				config.ChaosRedisAndMongo,
			)),
		//ava.Cors(lib.Cors()),
	)
}

var s6 = `[
  {
    "device_id": "6c3f4cb6c5899478efrgea",
    "functions": [
      {
        "values": {},
        "code": "switch_1",
        "type": "Boolean",
        "value_range_json": [
          [
            true,
            "开启"
          ],
          [
            false,
            "关闭"
          ]
        ]
      },
      {
        "values": {},
        "code": "switch_2",
        "type": "Boolean",
        "value_range_json": [
          [
            true,
            "开启"
          ],
          [
            false,
            "关闭"
          ]
        ]
      }
    ]
  }
]
`

func TestString(t *testing.T) {
	ava.Debug(removeWhitespace(s6))
}
func TestSS190(t *testing.T) {
	var resp interface{}
	err := tuyago.Post(ava.Background(), "/v1.0/aispeech/voice/push", &struct {
		Tts      string `json:"tts"`
		DeviceId string `json:"deviceId"`
	}{Tts: "您好，\n很高兴为您服务，我们正在处理您的问题，请耐心等待。", DeviceId: "6c06757f132f8977ca3yho"}, &resp)
	if err != nil {
		t.Fatal(err)
	}
	//
	//var resp1 interface{}
	//err = tuyago.Get(ava.Background(), fmt.Sprintf("/v1.0/homes/%s/rooms/%d/devices", "178176713", 53096637), &resp1)
	//
	//if err != nil {
	//	t.Fatal(err)
	//}
}

func TestConsumer(t *testing.T) {
	select {}
}

type grouInfo struct {
	SpaceId   string `json:"space_id"`
	GroupName string `json:"group_name"`
	ProductId string `json:"product_id"`
	DeviceIds string `json:"device_ids"`
}

func TestGroup(t *testing.T) {
	var result struct {
		Success bool `json:"success"`
	}

	var g = &grouInfo{
		SpaceId:   "178176713",
		GroupName: "测试",
		ProductId: "",
		DeviceIds: "",
	}

	err := tuyago.Post(ava.Background(), "/v2.0/cloud/thing/group", g, &result)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun(t *testing.T) {
	//Run(ava.Background(), "ay1716438065043jAiE1", "178176713", "打开客厅双色1号温明装射灯")
	//Run(ava.Background(), "ay1716438065043jAiE1", "178176713", "打开所有")
	//Run(ava.Background(), "ay1716438065043jAiE1", "178176713", "关闭客厅所有灯")
	//Run(ava.Background(), "ay1716438065043jAiE1", "178176713", "同步设备")
	//Run(ava.Background(), "ay1716438065043jAiE1", "178176713", "关闭所有灯")
	//Run(ava.Background(), "ay1716438065043jAiE1", "178176713", "关闭所有灯")
	//Run(ava.Background(), "ay1716438065043jAiE1", "178176713", "你好")
	//Run(ava.Background(), "ay1716438065043jAiE1", "178176713", "我上一句说了什么")
	//Run(ava.Background(), "ay1716438065043jAiE1", "178176713", "今天贵阳的天气怎么样")
	//Run(ava.Background(), "ay1716438065043jAiE1", "178176713", "调高地暖温度")
	//Run(ava.Background(), "ay1716438065043jAiE1", "178176713", "我有多少个设备")
	//Run(ava.Background(), "ay1716438065043jAiE1", "178176713", "我只有16个设备，为什么是20个")
	//Run(ava.Background(), "ay1716438065043jAiE1", "178176713", "灯都关闭了吗")
	//Run(ava.Background(), "ay1716438065043jAiE1", "178176713", "创建一个离家模式")
	//Run(ava.Background(), "ay1716438065043jAiE1", "178176713", "我想睡觉了")
	//Run(ava.Background(), "ay1716438065043jAiE1", "178176713", "你会去控制哪些设备")
	//Run(ava.Background(), "ay1716438065043jAiE1", "178176713", "执行回家场景，晚安")
	Run(ava.Background(), "ay1716438065043jAiE1", "178176713", "创建一个有人灯亮的自动化")

	//select {}
}

func TestPrompts(t *testing.T) {
	var tmp = `
{{.history}}
`
	p := prompts.NewPromptTemplate(tmp, []string{"history"})
	s, err := p.Format(map[string]any{"history": "ssss"})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(s)
}

/*
模型	描述	输出维度
text-embedding-3-large	最适合英语和非英语任务的嵌入模型	3,072
text-embedding-3-small	性能比第二代 ada 嵌入模型更高	1,536
text-embedding-ada-002	最强大的第二代嵌入模型，取代了 16 个第一代模型	1,536
*/
func TestEmbedding(t *testing.T) {

	opts := []openai.Option{
		//openai.WithEmbeddingModel("text-embedding-3-large"),
		openai.WithEmbeddingModel("text-embedding-v3"),
		openai.WithBaseURL(defaultUrl),
		openai.WithToken(defaultKey),
		//openai.WithModel("gpt-4o-mini-2024-07-18"),
		//openai.WithModel("qwen-turbo-latest"),
		//openai.WithResponseFormat(openai.ResponseFormatJSON),
		openai.WithCallback(LogHandler{}),
	}
	llm, err := openai.New(opts...)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	e, err := embeddings.NewEmbedder(llm)
	if err != nil {
		t.Fatal(err)
	}

	//embedings, err := llm.CreateEmbedding(ctx, []string{"ola", "mundo"})
	//if err != nil {
	//	t.Fatal(err)
	//}

	store, err := redisvector.New(ctx,
		redisvector.WithConnectionURL("redis://:ojo1QbOygiKjT1uZ@47.106.129.170:6379/0"),
		//redisvector.WithConnectionURL("47.106.129.170:6379"),
		redisvector.WithIndexName("test_redis_vectorstore", true),
		redisvector.WithEmbedder(e),
	)
	if err != nil {
		t.Fatal(err)
	}

	data := []schema.Document{
		{PageContent: "Tokyo", Metadata: map[string]any{"population": 9.7, "area": 622}},
		{PageContent: "Kyoto", Metadata: map[string]any{"population": 1.46, "area": 828}},
		{PageContent: "Hiroshima", Metadata: map[string]any{"population": 1.2, "area": 905}},
		{PageContent: "Kazuno", Metadata: map[string]any{"population": 0.04, "area": 707}},
		{PageContent: "Nagoya", Metadata: map[string]any{"population": 2.3, "area": 326}},
		{PageContent: "Toyota", Metadata: map[string]any{"population": 0.42, "area": 918}},
		{PageContent: "Fukuoka", Metadata: map[string]any{"population": 1.59, "area": 341}},
		{PageContent: "Paris", Metadata: map[string]any{"population": 11, "area": 105}},
		{PageContent: "London", Metadata: map[string]any{"population": 9.5, "area": 1572}},
		{PageContent: "Santiago", Metadata: map[string]any{"population": 6.9, "area": 641}},
		{PageContent: "Buenos Aires", Metadata: map[string]any{"population": 15.5, "area": 203}},
		{PageContent: "Rio de Janeiro", Metadata: map[string]any{"population": 13.7, "area": 1200}},
		{PageContent: "Sao Paulo", Metadata: map[string]any{"population": 22.6, "area": 1523}},
	}

	_, err = store.AddDocuments(ctx, data)
	docs, err := store.SimilaritySearch(ctx, "Tokyo的人口是多少", 2,
		vectorstores.WithScoreThreshold(0.1),
	)
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(docs)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(b))
}
