package miniprogram

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"vinesai/internel/ava"
	"vinesai/internel/config"
	"vinesai/internel/db"
	"vinesai/internel/db/db_hub"
	"vinesai/internel/x"
	"vinesai/proto/phub"
	"vinesai/proto/pmini"

	"github.com/sashabaranov/go-openai"
)

type hub struct {
	//user:chan
	room map[string]chan *pmini.ChatStreamRsp
	mux  *sync.RWMutex
	//用户发送的需要处理的数据
	handlerMessage chan *message
	isClose        sync.Map
}

type message struct {
	userId  string
	content string
}

var gHub *hub

func hubInstance() *hub {
	if gHub != nil {
		return gHub
	}

	gHub = &hub{
		mux:            new(sync.RWMutex),
		room:           make(map[string]chan *pmini.ChatStreamRsp, 1024),
		handlerMessage: make(chan *message, 10240),
	}

	go gHub.stream()
	return gHub
}

func (h *hub) addHub(userId string, ch chan *pmini.ChatStreamRsp) {
	h.mux.Lock()
	if h.room[userId] == nil {
		h.room[userId] = ch
		h.isClose.Store(userId, userId)
	}
	h.mux.Unlock()
}

func (h *hub) removeHub(userId string) {
	h.mux.Lock()
	if v := h.room[userId]; v != nil {
		delete(h.room, userId)
		h.isClose.Delete(userId)
	}
	h.mux.Unlock()
}

var OpenAi *openai.Client

func ChaosOpenAI() error {

	if config.GConfig.OpenAI.BaseURL != "" {
		ocf := openai.DefaultConfig(config.GConfig.OpenAI.Key)
		ocf.BaseURL = config.GConfig.OpenAI.BaseURL
		OpenAi = openai.NewClientWithConfig(ocf)
	} else {
		OpenAi = openai.NewClient(config.GConfig.OpenAI.Key)
	}
	//_, err := ask(ava.Background(), "你是一个得力的居家助手", "test")

	if OpenAi == nil {
		panic("openai.Client is nil")
	}

	return nil
}

func paramBuild(msg string, history []*phub.ChatHistory) []openai.ChatCompletionMessage {

	if len(history) > 3 {
		history = history[len(history)-3:]
	}

	var mesList = make([]openai.ChatCompletionMessage, 0, 6)

	//设置配置指令
	mesList = append(mesList, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: "你现在是一个私人助理",
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
	return mesList
}

func (h *hub) stream() {

	for msg := range h.handlerMessage {

		var m = msg

		h.mux.RLock()
		ch, ok := h.room[m.userId]
		if !ok {
			ava.Debugf("msg=%v |m=%v", x.MustMarshal2String(m), x.MustMarshal2String(msg))
			continue
		}
		h.mux.RUnlock()

		//发消息之前判断是否关闭
		if _, ok := h.isClose.Load(m.userId); !ok {
			continue
		}
		//先将消息发给自己
		ch <- &pmini.ChatStreamRsp{
			Code: http.StatusOK, //表示结束了
			Msg:  "success",
			Data: &pmini.ChatStreamData{
				Content:     m.content,
				UserId:      m.userId,
				Sender:      "YOU:",
				IsUser:      true,
				End:         true,
				DisplayName: true,
			},
		}

		go func() {
			//接收到用户提问请求，处理单条请求，包括发送给chatgpt
			//从数据库取出当前用户最近的3条记录,作为上下文
			var dbHistory []*db_hub.MessageHistory
			err := db.
				GMysql.
				Table(db_hub.TableMessageHistory).
				Where("identity=?", m.userId).
				Order("created_at desc").
				Limit(3).
				Find(&dbHistory).Error
			if err != nil {
				ava.Error(err)
				return
			}

			var history []*phub.ChatHistory
			for i := len(dbHistory) - 1; i >= 0; i-- {
				var d = &phub.ChatHistory{
					Message: dbHistory[i].Message,
					Resp:    dbHistory[i].Resp,
				}
				history = append(history, d)
			}

			mesList := paramBuild(m.content, history)

			stream, err := OpenAi.CreateChatCompletionStream(
				context.Background(),
				openai.ChatCompletionRequest{
					Model:     openai.GPT3Dot5Turbo,
					Messages:  mesList,
					MaxTokens: 2048,
					//Temperature: config.GConfig.OpenAI.Temperature,
					//TopP:        config.GConfig.OpenAI.TopP,
					//Temperature: 0.5,
					//TopP:        1,
					//N:           1,
					Stream: true,
				},
			)

			if err != nil {
				ava.Error(err)
				return
			}

			var msgStr strings.Builder

			for {
				response, err := stream.Recv()
				if err != nil {
					ava.Debugf("err=%v |data=%v", err, msgStr.String())
					//发消息之前判断是否关闭
					if _, ok := h.isClose.Load(m.userId); !ok {
						break
					}
					//通知前端结束了
					ch <- &pmini.ChatStreamRsp{
						Code: http.StatusOK, //表示结束了
						Msg:  "success",
						Data: &pmini.ChatStreamData{
							Content:     "",
							UserId:      m.userId,
							Sender:      "AI:",
							IsUser:      false,
							End:         true,
							DisplayName: false,
						},
					}
					break
				}

				if len(response.Choices) == 0 {
					continue
				}

				if response.Choices[0].Delta.Content == "" {
					continue
				}

				//发消息之前判断是否关闭
				if _, ok := h.isClose.Load(m.userId); !ok {
					break
				}

				var DisplayName bool
				if msgStr.Len() == 0 {
					DisplayName = true
				}

				ch <- &pmini.ChatStreamRsp{
					Code: http.StatusOK,
					Msg:  "success",
					Data: &pmini.ChatStreamData{
						Content:     response.Choices[0].Delta.Content,
						UserId:      m.userId,
						Sender:      "AI",
						IsUser:      false,
						End:         false,
						DisplayName: DisplayName,
					},
				}

				msgStr.WriteString(response.Choices[0].Delta.Content)

			}

			stream.Close()

			//消息入库
			var h = &db_hub.MessageHistory{
				Message:  m.content,
				Resp:     msgStr.String(),
				Option:   2,
				Identity: m.userId,
			}

			//消息入库
			err = db.GMysql.Table(db_hub.TableMessageHistory).Create(h).Error
			if err != nil {
				ava.Error(err)
			}
		}()
	}
}
