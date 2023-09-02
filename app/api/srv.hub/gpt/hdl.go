package gpt

import (
	"net/http"
	"vinesai/internel/ava"
	"vinesai/internel/x"
	"vinesai/proto/phub"
)

type Gpt struct{}

// 请求chatgpt
func (g *Gpt) Ask(c *ava.Context, req *phub.ChatReq, rsp *phub.ChatRsp) {
	if req.Message == "" || req.HomeId == "" {
		rsp.Code = http.StatusBadRequest
		rsp.Msg = x.StatusBadRequest
		return
	}

	r, err := ask(c, req.Message, req.HomeId)
	if err != nil {
		c.Error(err)
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = x.StatusInternalServerError
		return
	}
	var chatData = &phub.ChatData{
		Tip:  r.Tip,
		Exp:  r.Exp,
		Resp: r.Resp,
	}

	rsp.Code = http.StatusOK
	rsp.Msg = x.StatusOK
	rsp.Data = chatData
}
