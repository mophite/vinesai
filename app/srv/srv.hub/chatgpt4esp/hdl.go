package chatgpt4esp

import (
	"net/http"

	"vinesai/internel/ava"
	"vinesai/internel/config"
	"vinesai/internel/db/db_hub"
	"vinesai/internel/x"
	"vinesai/proto/phub"
)

type Gpt struct{}

// 请求chatgpt
func (g *Gpt) Ask(c *ava.Context, req *phub.ChatReq, rsp *phub.ChatRsp) {
	if req.Message == "" {
		rsp.Code = http.StatusBadRequest
		rsp.Msg = x.StatusBadRequest
		return
	}

	var r *db_hub.MessageHistory
	var err error
	switch config.GConfig.OpenAI.Method {
	case "1":
		r, err = methodOne(c, req)
	case "2":
		r, err = methodTwo(c, req)
	case "3":
		r, err = methodThree(c, req)
	}

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
