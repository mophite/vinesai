package ipc

import (
	"errors"
	"vinesai/internel/ava"
	"vinesai/proto/phub"
)

var chatClient phub.ChatClient

func InitIpc() {
	chatClient = phub.NewChatClient()
}

func Chat2AI(c *ava.Context, req *phub.ChatReq) (*phub.ChatRsp, error) {
	rsp, err := chatClient.Ask(c, req, ava.WithName("srv.hub"))
	if err != nil {
		return nil, err
	}

	if rsp.Data == nil {
		return nil, errors.New("no data")
	}

	return rsp, nil
}
