package homeassistant

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"
	"vinesai/internel/ava"
	"vinesai/internel/x"

	"github.com/gorilla/websocket"
)

func init() {
	newHub()

	var wait = new(sync.WaitGroup)

	for k, _ := range mapHome2Url {
		wait.Add(1)
		go websocketHa(wait, k)
	}

	wait.Wait()

	for k, _ := range mapHome2Url {
		go runXiaoMiSpeaker(k)
	}
}

// 用户对应的url每个用户都有不同的端口号,后期配置成域名
// 手动后台配置,123是家庭标识符,后期做授权
// 授权之后才能使用家庭的api
var mapHome2Url = map[string]string{
	"123": "127.0.0.1:8123",
}

var mapUserToken = map[string]string{
	"123": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJkOTQyYzVlYWI1ZWI0OWVhOWQxNzM2MjhhMDg0YWNlMyIsImlhdCI6MTcxNjM5MzU0OCwiZXhwIjoyMDMxNzUzNTQ4fQ.qqpUAAALF4-DQOsPT6sTuEXaZX8REjJkGM5rQF2bghY",
}

// 服务信息过滤
// todo 信息缓存,异步通知才更新
// 过滤掉的内容记得要打开
var filterServiceDomain = map[string]bool{
	"openai_conversation":     true,
	"remote":                  true,
	"update":                  true,
	"text":                    true,
	"number":                  true,
	"person":                  true,
	"homeassistant":           true,
	"persistent_notification": true,
	"system_log":              true,
	"logger":                  true,
	"recorder":                true,
	"frontend":                true,
	"cloud":                   true,
	"ffmpeg":                  true,
	"tts":                     true,
	"scene":                   true,
	"input_number":            true,
	"logbook":                 true,
	"input_select":            true,
	"input_button":            true,
	"entity":                  true,
	"input_boolean":           true,
	"script":                  true,
	"zone":                    true,
	"conversation":            true,
	"schedule":                true,
	"backup":                  true,
	"input_datetime":          true,
	"xiaomi_miot":             true,
	"input_text":              true,
	"counter":                 true,
	"notify":                  true,
	"device_tracker":          true,
	"climate":                 true, //气候
	"automation":              true,
}

func getServices(c *ava.Context, home string) (string, error) {
	req, _ := http.NewRequest("GET", "http://"+mapHome2Url[home]+"/api/services", nil)
	req.Header.Set("Authorization", mapUserToken[home])
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.Error(err)
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	c.Info("--------------services", x.BytesToString(body))

	var filter = make([]map[string]interface{}, 0, 100)
	err = x.MustUnmarshal(body, &filter)
	if err != nil {
		c.Error(err)
		return "", err
	}

	var result = make([]map[string]interface{}, 0, 50)

	//执行过滤
	for i := range filter {
		for k, v := range filter[i] {
			if k == "domain" && !filterServiceDomain[v.(string)] {
				result = append(result, filter[i])
				break
			}
		}
	}

	return x.MustMarshal2String(result), nil

}

// 设备信息过滤
// todo 实时获取
var filterState = map[string]bool{
	"person.":    true,
	"zone.home":  true,
	"iphone":     true,
	"sun":        true,
	"update":     true,
	"automation": true,
}

type shortStates struct {
	EntityId   string `json:"entity_id"`
	Attributes struct {
		FriendlyName string `json:"friendly_name"`
	} `json:"attributes"`
}

func getStates(c *ava.Context, home string) (string, []*shortStates, error) {
	req, _ := http.NewRequest("GET", "http://"+mapHome2Url[home]+"/api/states", nil)
	req.Header.Set("Authorization", mapUserToken[home])
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.Error(err)
		return "", nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	c.Info("-----------------", x.BytesToString(body))

	var filter = make([]map[string]interface{}, 0, 100)
	err = x.MustUnmarshal(body, &filter)
	if err != nil {
		c.Error(err)
		return "", nil, err
	}

	var ss []*shortStates
	err = x.MustUnmarshal(body, &ss)
	if err != nil {
		c.Error(err)
	}
	//fmt.Println("----------", x.MustMarshal2String(ss))

	var result = make([]map[string]interface{}, 0, 50)

	//执行过滤
	for i := range filter {
	Out:
		for k, v := range filter[i] {
			if strings.Contains(k, "entity_id") {
				for k1, _ := range filterState {
					if strings.Contains(v.(string), k1) {
						break Out
					}
				}

				result = append(result, filter[i])
				break
			}
		}
	}

	return x.MustMarshal2String(result), ss, nil

}

var (
	httpClient *http.Client
)

func init() {
	tr := &http.Transport{
		//Proxy:           http.ProxyFromEnvironment,
		//TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second, //拨号等待连接完成的最大时间
			KeepAlive: 30 * time.Second, //保持网络连接活跃keep-alive探测间隔时间。
		}).DialContext,
		MaxIdleConns:        200,
		IdleConnTimeout:     300 * time.Second,
		MaxIdleConnsPerHost: 200,
	}
	httpClient = &http.Client{
		Transport: tr,
		Timeout:   30 * time.Second, //设置超时，包含connection时间、任意重定向时间、读取response body时间
	}
}

func callServiceWs(home string, data interface{}) {
	ava.Debugf("callServiceHttpWs |home=%s |data=%s", home, x.MustMarshal2String(data))
	gHub.writeJson(home, data)
}

func callServiceHttp(home, service string, data []byte) {
	ava.Debugf("callServiceHttp |url=%s |data=%s", "http://"+mapHome2Url[home]+"/api/services/"+service, string(data))
	req, _ := http.NewRequest("POST", "http://"+mapHome2Url[home]+"/api/services/"+service, bytes.NewReader(data))
	req.Header.Set("Authorization", mapUserToken[home])
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		ava.Error(err)
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ava.Error(err)
		return
	}

	ava.Infof("callServiceHttp |body=%s", x.BytesToString(body))
}

var aiTmp = `你是一个home-assistant智能家居管家，当我想要控制智能家居设备时，通过friendly_name去判断设备位置。
我给你提供的数据都是通过homeassistant这个开源项目的REST API得到的。
例如：1.将【客厅led】改为黄色：
【设备清单】：
[{
"entity_id": "light.smart_led_strip_2",
"state": "on",
"attributes": {
"supported_color_modes": [
"hs"
],
"color_mode": "hs",
"brightness": 255,
"hs_color": [
9,
67.1
],
"rgb_color": [
255,
109,
83
],
"xy_color": [
0.591,
0.328
],
"friendly_name": "客厅led",
"supported_features": 0
},
"last_changed": "2024-05-27T07:10:25.132514+00:00",
"last_reported": "2024-05-27T07:10:36.474833+00:00",
"last_updated": "2024-05-27T07:10:36.473066+00:00",
"context": {
"id": "01HYWE60BB4Q3X0ZM1B43ZVV9Z",
"parent_id": null,
"user_id": "94b4baac240646069282442afdb582d0"
}
}]

【指令清单】：
[{
  "domain": "light",
  "services": {
    "turn_on": {
      "name": "Turn on",
      "fields": {
        "rgb_color": {
          "filter": {
            "attribute": {
              "supported_color_modes": [
                "hs",
                "xy",
                "rgb",
                "rgbw",
                "rgbww"
              ]
            }
          },
          "selector": {
            "color_rgb": null
          },
          "example": "[255, 100, 100]",
          "name": "Color",
          "description": "The color in RGB format. A list of three integers between 0 and 255 representing the values of red, green, and blue."
        }
      }
    }
  }
}]

【指令结果】：
{"data":{"entity_id":"light.smart_led_strip","rgb_color": [255, 255, 0]},"service":"light/turn_on","message":"好的主人,已为你修改客厅led颜色"}
其中：
"service":通过【指令清单】"domain"的light加上services指令集"turn_on"得到
"data":通过设备当前设备状态修改后的数据
"message":以俏皮的语气回答我，并告诉我结果。

2.关闭【客厅led】：
【设备清单】：
[{
"entity_id": "light.smart_led_strip_2",
"state": "on",
"attributes": {
"supported_color_modes": [
"hs"
],
"color_mode": "hs",
"brightness": 255,
"hs_color": [
9,
67.1
],
"rgb_color": [
255,
109,
83
],
"xy_color": [
0.591,
0.328
],
"friendly_name": "客厅led",
"supported_features": 0
},
"last_changed": "2024-05-27T07:10:25.132514+00:00",
"last_reported": "2024-05-27T07:10:36.474833+00:00",
"last_updated": "2024-05-27T07:10:36.473066+00:00",
"context": {
"id": "01HYWE60BB4Q3X0ZM1B43ZVV9Z",
"parent_id": null,
"user_id": "94b4baac240646069282442afdb582d0"
}
}]

【指令清单】：
[{
  "domain": "light",
  "services": {
    "turn_off": {
      "name": "Turn off"
    }
  }
}]

【指令结果】：
{"data":{"entity_id":"light.smart_led_strip_2"},"service":"light/turn_off","message":"好的主人"}

当前所有指令清单和设备清单如下：
【指令清单】：%s
【设备清单】：%s

当我向你发起会话，告诉你我的意图，如果是需要控制家居设备的时候，把【指令结果】按照JSON数据格式发给我，{}前后不要有任何内容，不然我无法解析。
`

var aiTmp1 = `你是一个智能家居助手,负责通过 Home Assistant REST API 控制家中的智能设备。请根据我提供的指令和设备清单,执行相应的操作,并以 JSON 格式返回指令结果。
在判断设备位置时,请使用 friendly_name 字段，还要注意state的状态是否可用，指令结果中包含以下字段:
service:结合指令清单中的 domain 和 services 得到,如 "light/turn_on"。
data:根据设备当前状态和指令要求,生成修改后的设备数据。 
message:以友好、俏皮的口吻告知操作结果。 
例如:
1.将【客厅led】改为黄色:
{"serviceData":[{"data":{"entity_id":"light.smart_led_strip_2","rgb_color":[255,255,0]},"service":"light/turn_on"}],"message":"好的主人，已将客厅led改为黄色"}
2.关闭【客厅led】:
{"serviceData":[{"data":{"entity_id":"light.smart_led_strip_2"},"service":"light/turn_off"}],"message":"客厅led已乖乖睡觉啦~"}
3.打开【客厅插座】:
{"serviceData":[{"data":{"entity_id":"switch.qmi_psv3_4067_switch"},"service":"switch/turn_on"}],"message":"客厅插座已打开"}
请直接给出 JSON 格式的指令结果,不要有其他文字。当前的指令清单和设备清单如下: 【指令清单】:%s 【设备清单】:%s`

var aiTmp2 = `你是一个理解home-assistant REST API的智能家居助手，我将为您提供此任务的基础知识，请在之后使用它来完成任务。
<services></services>表示调用的服务，<states></states>表示设备清单，<notice></notice>表示设备清单数据使用注意事项，非常重要。

<services>
%s
</services>

<states>
%s
</states>

<notice>
- 在<states>中，friendly_name(只能通过这个字段去识别设备名称和设备位置)，state(设备状态,on表示打开，off表示关闭，unavailable表示不可用,unavailable状态的设备你无法控制)
- {"entity_id": "select.smart_plug_power_on_behavior","state":"unavailable"}表示设备不可用，这个时候直接告诉我设备发生故障即可
</notice>

请根据我的意图，选择合适的设备和指令执行相应的操作,并以 JSON 格式返回指令结果。
<example>
1.将【客厅led】改为黄色:
{"serviceData":[{"data":{"entity_id":"light.smart_led_strip_2","rgb_color":[255,255,0]},"domain":"light","service":"turn_on"}],"message":"好的主人，已将客厅led改为黄色"}
2.关闭【客厅led】:
{"serviceData":[{"data":{"entity_id":"light.smart_led_strip_2"},"domain":"light","service":"turn_off"}],"message":"客厅led已乖乖睡觉啦~"}
3.打开【客厅插座】:
{"serviceData":[{"data":{"entity_id":"switch.qmi_psv3_4067_switch"},"domain":"switch","service":"turn_on"}],"message":"客厅插座已打开"}
4.不可用设备:
{"serviceData":[{"data":{"entity_id":"switch.smart_plug_socket_1","state":"unavailable"},"domain":"","service":""}],"message":"卧室插座不可用"}

- service:根据<services>中的domain和services 得到请求home-assistant的服务,如 "light/turn_on" 。
- data:你要修改的设备状态数据，其中entity_id将要修改的设备实体标识 。
- message:以友好、俏皮的口吻告知修改结果。
</example>`

var aiTmp2Ws = `
<|im_start|>system
你是一个理解home assistant REST API的智能家居助手，可以控制房子里的设备。按照指示完成以下任务或仅使用提供的信息回答以下问题。

控制指令(/api/services/{{domain}}/{{services}}：%s

设备列表(/api/states)：%s

注意事项：
1.在<states>中，friendly_name(只能通过这个字段去识别设备名称和设备位置)；
2.{"entity_id": "select.smart_plug_power_on_behavior","state":"unavailable"}表示设备不可用，直接汇报设备故障。

请根据我的意图，选择合适的设备和指令执行相应的操作,并以 JSON 格式返回指令结果。
1.将【客厅led】改为黄色:
{"serviceData":[{"type":"call_service","domain":"light","service":"turn_on","service_data":{"rgb_color":[255,255,0]},"target":{"entity_id":"light.smart_led_strip_2"}}],"message":"好的主人，已将客厅led改为黄色"}
2.关闭【客厅led】:
{"serviceData":[{"type":"call_service","domain":"light","service":"turn_off","service_data":{},"target":{"entity_id":"light.smart_led_strip_2"}}],"message":"客厅led已乖乖睡觉啦~"}
3.打开【客厅插座】:
{"serviceData":[{"type":"call_service","domain":"switch","service":"turn_on","service_data":{},"target":{"entity_id":"switch.qmi_psv3_4067_switch"}}],"message":"客厅插座已打开"}
4.不可用设备:
{"serviceData":[{"target":{"entity_id":"light.smart_led_strip_2"}}],"message":"卧室插座不可用"}

- domain:要调用的指令服务,例如：light表示灯，switch表示开关。
- service:服务指令的具体，例如：turn_on表示打开，turn_off表示关闭 。
- service_data:你要修改的设备状态数据
- target:目标，entity_id表示设备实体唯一标识 。
- message:以友好、俏皮的口吻告知修改结果。
<|im_end|>`

//- 【设备清单】中如果是{ "entity_id": "switch.smart_plug_socket_1", "state": "unavailable"},则提示我“xxx不可用”,参考第4点。

type hub struct {
	connLock *sync.RWMutex
	conn     map[string]*websocket.Conn

	//记录实体
	entityLock *sync.RWMutex
	entity     map[string]*entity
}

var gHub *hub

func newHub() {
	gHub = &hub{
		connLock:   new(sync.RWMutex),
		conn:       make(map[string]*websocket.Conn, 50),
		entityLock: new(sync.RWMutex),
		entity:     make(map[string]*entity, 50),
	}
}

func (h *hub) getConn(home string) *websocket.Conn {
	h.connLock.RLock()
	defer h.connLock.RUnlock()
	return h.conn[home]
}

func (h *hub) addConn(home string, conn *websocket.Conn) {
	h.connLock.Lock()
	h.conn[home] = conn
	h.connLock.Unlock()
}

func (h *hub) removeConn(home string) {
	h.connLock.Lock()
	delete(h.conn, home)
	h.connLock.Unlock()
}

func (h *hub) getEntity(home string) *entity {
	h.entityLock.RLock()
	defer h.entityLock.RUnlock()
	return h.entity[home]
}

func (h *hub) addEntity(home string, entity *entity) {
	h.entityLock.Lock()
	h.entity[home] = entity
	h.entityLock.Unlock()
}

func (h *hub) removeEntity(home string) {
	h.entityLock.Lock()
	delete(h.entity, home)
	h.entityLock.Unlock()
}

func (h *hub) writeJson(home string, data interface{}) {
	h.connLock.RLock()
	defer h.connLock.RUnlock()
	c, ok := h.conn[home]
	if !ok {
		return
	}

	c.WriteJSON(data)
}

type stateData struct {
	Type  string `json:"type"`
	Event struct {
		EventType string `json:"event_type"`
		Data      struct {
			NewState struct {
				EntityID string `json:"entity_id"`
				State    string `json:"state"` //语音内容
			} `json:"new_state"`
		} `json:"data"`
	} `json:"event"`
}

func websocketHa(wait *sync.WaitGroup, home string) {

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	var host, accessToken = mapHome2Url[home], mapUserToken[home]

	accessToken = strings.TrimPrefix(accessToken, "Bearer ")

	u := url.URL{Scheme: "ws", Host: host, Path: "/api/websocket"}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		ava.Errorf("host=%s |token=%s |err=%v", host, accessToken, err)
		return
	}

	defer conn.Close()

	//过滤掉要求
	conn.ReadMessage()

	//鉴权
	var authReq = struct {
		Type        string `json:"type"`
		AccessToken string `json:"access_token"`
	}{Type: "auth", AccessToken: accessToken}

	err = conn.WriteJSON(&authReq)
	if err != nil {
		ava.Errorf("host=%s |token=%s |err=%v", host, accessToken, err)
		return
	}

	_, message, err := conn.ReadMessage()
	if err != nil {
		ava.Errorf("host=%s |token=%s |err=%v", host, accessToken, err)
		return
	}

	ava.Debugf("auth |message=%s", string(message))

	type Result struct {
		Type    string `json:"type"`
		Success bool   `json:"success"`
	}
	var result Result

	err = x.MustUnmarshal(message, &result)
	if err != nil {
		ava.Errorf("host=%s |token=%s |err=%v", host, accessToken, err)
		return
	}

	if result.Type != "auth_ok" {
		ava.Errorf("host=%s |token=%s", host, accessToken)
		return
	}

	ava.Debugf("handshake suscess |host=%s |token=%s", host, accessToken)

	//监听状态变化
	var state = struct {
		Id        int    `json:"id"`
		Type      string `json:"type"`
		EventType string `json:"event_type"`
	}{Id: ava.RandInt(1, 100000), Type: "subscribe_events", EventType: "state_changed"}

	err = conn.WriteJSON(&state)
	if err != nil {
		ava.Errorf("host=%s |token=%s |err=%v", host, accessToken, err)
		return
	}

	_, stateMessage, err := conn.ReadMessage()
	if err != nil {
		ava.Errorf("host=%s |token=%s |err=%v", host, accessToken, err)
		return
	}

	ava.Debugf("state_changed |message=%s", string(stateMessage))

	var stateResult Result

	err = x.MustUnmarshal(stateMessage, &stateResult)
	if err != nil {
		ava.Errorf("host=%s |token=%s |err=%v", host, accessToken, err)
		return
	}

	if !stateResult.Success {
		ava.Errorf("host=%s |token=%s |stateResult=%v", host, accessToken, stateResult)
		return
	}

	gHub.addConn(home, conn)

	wait.Done()

	defer gHub.removeConn(home)

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				ava.Error(err)
				return
			}

			var fromState stateData
			err = x.MustUnmarshal(message, &fromState)
			if err != nil {
				ava.Error(err)
				continue
			}

			if isXiaoMiSpeaker(fromState.Event.Data.NewState.EntityID) {
				err = recevieMessage(home, fromState.Event.Data.NewState.State)
				if err != nil {
					ava.Error(err)
					continue
				}
			}

		}
	}()

	// todo {"id":40,"type":"result","success":false,"error":{"code":"id_reuse","message":"Identifier values have to increase."}}
	//var idIncrease = ava.RandInt32(1, 100)
	//
	////发送心跳包
	//var quit = make(chan string)
	//x.TimingwheelTicker(time.Second*5, func() {
	//	err := conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{ "id": %d, "type": "ping" }`, atomic.AddInt32(&idIncrease, 1))))
	//	if err != nil {
	//		ava.Error(err)
	//	}
	//
	//	<-quit
	//})

	//退出
	for {
		select {
		//case <-quit:
		//
		//	// Cleanly close the connection by sending a close message and then
		//	// waiting (with timeout) for the server to close the connection.
		//	err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		//	if err != nil {
		//		ava.Error(err)
		//		return
		//	}
		//	select {
		//	case <-done:
		//	case <-time.After(time.Second):
		//	}
		//	return
		case <-done:
			return
		case <-interrupt:

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				ava.Error(err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
