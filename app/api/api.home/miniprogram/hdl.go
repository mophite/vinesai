package miniprogram

import (
	"net/http"
	"vinesai/internel/ava"
	"vinesai/internel/x"
	"vinesai/proto/pmini"
)

// 小程序人工智能
type Mini struct {
	hub *hub
}

func NewMini() *Mini {
	return &Mini{hub: hubInstance()}
}

func (m *Mini) Chat(c *ava.Context, req *pmini.ChatReq, rsp *pmini.ChatRsp) {

	if req.Content == "" {
		rsp.Code = http.StatusBadRequest
		rsp.Msg = "content消息内容不能为空"
		return
	}

	m.hub.handlerMessage <- &message{
		userId:  req.UserId,
		content: req.Content,
	}

	rsp.Code = http.StatusOK
	rsp.Msg = x.StatusOK
}

func (m *Mini) ChatStream(c *ava.Context, req *pmini.ChatStreamReq, exit chan struct{}) chan *pmini.ChatStreamRsp {

	var rsp = make(chan *pmini.ChatStreamRsp)

	m.hub.addHub(req.UserId, rsp)

	go func() {
	QUIT:
		for {
			select {
			case <-exit:
				break QUIT
			}
		}

		m.hub.removeHub(req.UserId)

		close(rsp)
	}()

	return rsp
}
