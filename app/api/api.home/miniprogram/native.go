package miniprogram

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"
	"vinesai/internel/ava"
	"vinesai/proto/pmini"

	"github.com/gorilla/websocket"
)

// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// 在这里检查请求的来源，决定是否允许升级为 WebSocket 连接
		// 你可以根据自己的需求进行逻辑处理
		return true // 允许所有来源
	},
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *hub
	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	rsp chan *pmini.ChatStreamRsp

	close chan struct{}

	userId string
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.conn.Close()
		c.hub.removeHub(c.userId)
		c.close <- struct{}{}
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		//打印收到的消息
		ava.Debugf("from |req=%v", string(data))

		data = bytes.TrimSpace(bytes.Replace(data, newline, space, -1))

		var req pmini.ChatReq
		err = ava.Unmarshal(data, &req)
		if err != nil {
			ava.Error(err)
			continue
		}

		c.hub.handlerMessage <- &message{
			userId:  req.UserId,
			content: req.Content,
		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case data, ok := <-c.rsp:
			if !ok {
				return
			}
			ava.Debugf("to |rsp=%v", data.Data.Content)

			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				fmt.Println("----writePump done----")
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				fmt.Println("----2----", err)
				return
			}

			send := ava.MustMarshal(data)

			w.Write(send)

			// Add queued chat messages to the current websocket message.
			//提升性能，但是如果数据量过大发送不及时，会断开连接
			//n := len(c.rsp)
			//for i := 0; i < n; i++ {
			//	w.Write(newline)
			//
			//	tmp := <-c.rsp
			//	s := ava.MustMarshal(tmp)
			//
			//	w.Write(s)
			//}

			if err := w.Close(); err != nil {
				fmt.Println("----3----", err)
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-c.close:
			return
		}
	}
}

// serveWs handles websocket requests from the peer.
func ServeWs(w http.ResponseWriter, r *http.Request) {

	userId := r.URL.Query().Get("userId")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		hub:    hubInstance(),
		conn:   conn,
		rsp:    make(chan *pmini.ChatStreamRsp, 256),
		close:  make(chan struct{}),
		userId: userId,
	}

	client.hub.addHub(userId, client.rsp)

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
