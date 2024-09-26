package homeassistant

import (
	"bytes"
	"fmt"
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

	for k := range mapHome2Url {
		go websocketHa(k)
	}

	time.Sleep(time.Second)

	for k := range mapHome2Url {
		var count = 1
		for {
			//判断连接是否存在
			c := gHub.getConn(k)
			if c == nil && count < 5 {
				count++
				time.Sleep(time.Second * time.Duration(count))
				continue
			}

			if count >= 5 && c == nil {
				panic("too many times to connect")
			}

			go runXiaoMiSpeaker(k)

			break
		}
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
	"input_boolean":           true,
	"script":                  true,
	"zone":                    true,
	"conversation":            true,
	"schedule":                true,
	"backup":                  true,
	"input_datetime":          true,
	"input_text":              true,
	"counter":                 true,
	"notify":                  true,
	"automation":              true,
	"timer":                   true,
	"cover":                   true, //窗帘之类
	"humidifier":              true, //加湿器
	"camera":                  true, //相机
	"vacuum":                  true, //吸尘器
	"water_heater":            true, //热水器
	"alarm_control_panel":     true,
	//"select":                  true,//空调操作会用到
	"siren":          true,
	"device_tracker": true, //设备追踪，暂时用不到
	"button":         true, //实体自带指令
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

	return string(body), nil

	//c.Info("--------------services", x.BytesToString(body))

	//var filter = make([]map[string]interface{}, 0, 100)
	//err = x.MustUnmarshal(body, &filter)
	//if err != nil {
	//	c.Error(err)
	//	return "", err
	//}

	//var result = make([]map[string]interface{}, 0, 50)
	//
	////执行过滤
	//for i := range filter {
	//	for k, v := range filter[i] {
	//		if k == "domain" && !filterServiceDomain[v.(string)] {
	//			result = append(result, filter[i])
	//			break
	//		}
	//	}
	//}

	//var shortData shortServicesStruct
	//err = x.MustUnmarshal(body, &shortData)
	//if err != nil {
	//	c.Error(err)
	//	return "", err
	//}

	//c.Info("--------------services", x.MustMarshal2String(shortData))

	//return x.MustMarshal2String(shortData), nil

}

// 设备信息过滤
// todo 实时获取
var filterState = map[string]bool{
	"person.":                     true,
	"zone.home":                   true,
	"iphone":                      true,
	"ipad":                        true,
	"sun":                         true,
	"update":                      true,
	"automation":                  true,
	"conversation.home_assistant": true,
	"huawei":                      true,
}

type shortStates struct {
	EntityId   string `json:"entity_id"`
	State      string `json:"state"`
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

	//c.Info("-----------------所有实体", x.BytesToString(body))
	//
	//var filter = make([]map[string]interface{}, 0, 100)
	//err = x.MustUnmarshal(body, &filter)
	//if err != nil {
	//	c.Error(err)
	//	return "", nil, err
	//}

	var ss []*shortStates
	err = x.MustUnmarshal(body, &ss)
	if err != nil {
		c.Error(err)
	}
	//fmt.Println("----------所有实体", x.MustMarshal2String(ss))

	//var result = make([]map[string]interface{}, 0, 50)

	var result = make([]*shortStates, 0, 30)

	//执行过滤
	//todo 暂时使用精简数据
	for i := range ss {
		v := ss[i]
		var exist bool
		for k1 := range filterState {
			if strings.Contains(v.EntityId, k1) {
				exist = true
				break
			}
		}

		if !exist {
			result = append(result, &(*v))
		}

	}

	//for i := range filter {
	//Out:
	//	for k, v := range filter[i] {
	//		if strings.Contains(k, "entity_id") {
	//			for k1, _ := range filterState {
	//				if strings.Contains(v.(string), k1) {
	//					break Out
	//				}
	//			}
	//
	//			result = append(result, filter[i])
	//			break
	//		}
	//	}
	//}

	c.Info("-----------------entity", x.MustMarshal2String(result))

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

func callServiceWs(id int64, home string, data command) {
	ava.Debugf("callServiceHttpWs |home=%s |data=%s", home, x.MustMarshal2String(data))

	if data.Data == nil {
		data.Data = struct{}{}
	}

	var to wsCallService
	to.Id = id
	to.Type = "call_service"
	to.ServiceData = data.Data
	to.Target.EntityId = data.EntityId

	service := strings.Split(data.Service, ".")
	if len(service) != 2 {
		ava.Errorf("not enough value |data=%v", data)
		return
	}

	to.Domain = service[0]
	to.Service = service[1]

	//to.ReturnResponse = true

	gHub.writeJson(home, &to)
}

func callServiceHttp(home, serviceDot string, data []byte) {
	list := strings.Split(serviceDot, ".")
	if len(list) != 2 {
		ava.Errorf("not enough value |data=%v", data)
		return
	}

	domain := list[0]
	service := list[1]

	ava.Debugf("callServiceHttp |url=%s |data=%s", "http://"+mapHome2Url[home]+"/api/services/"+domain+"/"+service, string(data))
	req, _ := http.NewRequest("POST", "http://"+mapHome2Url[home]+"/api/services/"+domain+"/"+service, bytes.NewReader(data))
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

var aiTmp = `
<|im_start|>system
你是一个理解home assistant REST API的智能家居助手，可以控制房子里的设备。按照指示完成以下任务或仅使用提供的信息回答以下问题。

所有的控制指令：%s

所有的设备列表：%s

注意事项：
1.在设备列表中，friendly_name(只能通过这个字段去识别设备名称和设备位置)；
2.在设备列表中，{"entity_id": "select.smart_plug_power_on_behavior","state":"unavailable"}表示设备不可用，直接汇报设备故障。

请根据我的意图，选择合适的设备和指令执行相应的操作,并以 JSON 格式返回指令结果。
1.将【客厅led】改为黄色:
{"commands":[{"entity_id":"light.smart_led_strip_2","service":"light.turn_on","data":{"rgb_color":[255,255,0]}}],"answer":"好的主人，已将客厅led改为黄色"}
2.关闭【客厅led】:
{"commands":[{"entity_id":"light.smart_led_strip_2","service":"light.turn_off","data":{}}],"answer":"客厅led已乖乖睡觉啦~"}
3.打开【客厅插座】:
{"commands":[{"entity_id":"switch.qmi_psv3_4067_switch","service":"switch.turn_on","data":{}}],"answer":"客厅插座已打开"}
4.不可用设备:
{"commands":[{"entity_id":"light.smart_led_strip_2"}],"answer":"卧室插座不可用"}

注意事项：
1.message:以友好、俏皮的口吻告知修改结果。
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
				EntityID     string `json:"entity_id"`
				State        string `json:"state"` //语音内容
				LastChanged  string `json:"last_changed"`
				LastReported string `json:"last_reported"`
				LastUpdated  string `json:"last_updated"`
				Attributes   struct {
					Timestamp    string `json:"timestamp"`
					DeviceClass  string `json:"device_class"`  //motion运动传感器
					FriendlyName string `json:"friendly_name"` //设备名称
				} `json:"attributes"`
			} `json:"new_state"`
		} `json:"data"`
		TimeFired string `json:"time_fired"`
	} `json:"event"`
}

//func websocketHa(home string) {
//
//	interrupt := make(chan os.Signal, 1)
//	signal.Notify(interrupt, os.Interrupt)
//
//	var host, accessToken = mapHome2Url[home], mapUserToken[home]
//
//	accessToken = strings.TrimPrefix(accessToken, "Bearer ")
//
//	u := url.URL{Scheme: "ws", Host: host, Path: "/api/websocket"}
//	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
//	if err != nil {
//		ava.Errorf("host=%s |token=%s |err=%v", host, accessToken, err)
//		return
//	}
//
//	defer conn.Close()
//
//	//过滤掉要求
//	conn.ReadMessage()
//
//	//鉴权
//	var authReq = struct {
//		Type        string `json:"type"`
//		AccessToken string `json:"access_token"`
//	}{Type: "auth", AccessToken: accessToken}
//
//	err = conn.WriteJSON(&authReq)
//	if err != nil {
//		ava.Errorf("host=%s |token=%s |err=%v", host, accessToken, err)
//		return
//	}
//
//	_, message, err := conn.ReadMessage()
//	if err != nil {
//		ava.Errorf("host=%s |token=%s |err=%v", host, accessToken, err)
//		return
//	}
//
//	ava.Debugf("auth |message=%s", string(message))
//
//	type Result struct {
//		Type    string `json:"type"`
//		Success bool   `json:"success"`
//	}
//	var result Result
//
//	err = x.MustUnmarshal(message, &result)
//	if err != nil {
//		ava.Errorf("host=%s |token=%s |err=%v", host, accessToken, err)
//		return
//	}
//
//	if result.Type != "auth_ok" {
//		ava.Errorf("host=%s |token=%s", host, accessToken)
//		return
//	}
//
//	ava.Debugf("handshake suscess |host=%s |token=%s", host, accessToken)
//
//	//监听状态变化
//	var state = struct {
//		Id        int    `json:"id"`
//		Type      string `json:"type"`
//		EventType string `json:"event_type"`
//	}{Id: ava.RandInt(1, 100000), Type: "subscribe_events", EventType: "state_changed"}
//
//	err = conn.WriteJSON(&state)
//	if err != nil {
//		ava.Errorf("host=%s |token=%s |err=%v", host, accessToken, err)
//		return
//	}
//
//	_, stateMessage, err := conn.ReadMessage()
//	if err != nil {
//		ava.Errorf("host=%s |token=%s |err=%v", host, accessToken, err)
//		return
//	}
//
//	ava.Debugf("state_changed |message=%s", string(stateMessage))
//
//	var stateResult Result
//
//	err = x.MustUnmarshal(stateMessage, &stateResult)
//	if err != nil {
//		ava.Errorf("host=%s |token=%s |err=%v", host, accessToken, err)
//		return
//	}
//
//	if !stateResult.Success {
//		ava.Errorf("host=%s |token=%s |stateResult=%v", host, accessToken, stateResult)
//		return
//	}
//
//	gHub.addConn(home, conn)
//
//	defer gHub.removeConn(home)
//
//	done := make(chan struct{})
//
//	go func() {
//		defer close(done)
//		for {
//			_, message, err := conn.ReadMessage()
//			if err != nil {
//				ava.Errorf("home=%s |err=%v", home, err)
//				return
//			}
//
//			c := ava.Background()
//			c.Debug("----------设备状态变更---------", x.BytesToString(message))
//
//			var fromState stateData
//			err = x.MustUnmarshal(message, &fromState)
//			if err != nil {
//				c.Error(err)
//				continue
//			}
//
//			if fromState.Event.Data.NewState.State == "unavailable" {
//				continue
//			}
//
//			if isXiaoMiSpeaker(fromState.Event.Data.NewState.EntityID) {
//				err = receiveMessage(c, home, fromState.Event.Data.NewState.State, fromState.Event.Data.NewState.State)
//				if err != nil {
//					c.Error(err)
//					continue
//				}
//			}
//		}
//	}()
//
//	// todo {"id":40,"type":"result","success":false,"error":{"code":"id_reuse","message":"Identifier values have to increase."}}
//	//var idIncrease = ava.RandInt32(1, 100)
//	//
//	////发送心跳包
//	//var quit = make(chan string)
//	//x.TimingwheelTicker(time.Second*5, func() {
//	//	err := conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{ "id": %d, "type": "ping" }`, atomic.AddInt32(&idIncrease, 1))))
//	//	if err != nil {
//	//		ava.Error(err)
//	//	}
//	//
//	//	<-quit
//	//})
//
//	//退出
//	for {
//		select {
//		//case <-quit:
//		//
//		//	// Cleanly close the connection by sending a close message and then
//		//	// waiting (with timeout) for the server to close the connection.
//		//	err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
//		//	if err != nil {
//		//		ava.Error(err)
//		//		return
//		//	}
//		//	select {
//		//	case <-done:
//		//	case <-time.After(time.Second):
//		//	}
//		//	return
//		case <-done:
//			return
//		case <-interrupt:
//
//			// Cleanly close the connection by sending a close message and then
//			// waiting (with timeout) for the server to close the connection.
//			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
//			if err != nil {
//				ava.Error(err)
//				return
//			}
//			select {
//			case <-done:
//			case <-time.After(time.Second):
//			}
//			return
//		}
//	}
//}

func websocketHa(home string) {

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	reconnect := func() (*websocket.Conn, error) {

		var host, accessToken = mapHome2Url[home], mapUserToken[home]
		accessToken = strings.TrimPrefix(accessToken, "Bearer ")

		u := url.URL{Scheme: "ws", Host: host, Path: "/api/websocket"}
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			ava.Errorf("host=%s |token=%s |err=%v", host, accessToken, err)
			return nil, err
		}

		//过滤掉要求
		conn.ReadMessage()

		// 鉴权
		var authReq = struct {
			Type        string `json:"type"`
			AccessToken string `json:"access_token"`
		}{Type: "auth", AccessToken: accessToken}

		err = conn.WriteJSON(&authReq)
		if err != nil {
			ava.Errorf("host=%s |token=%s |err=%v", host, accessToken, err)
			return nil, err
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			ava.Errorf("host=%s |token=%s |err=%v", host, accessToken, err)
			return nil, err
		}

		type Result struct {
			Type    string `json:"type"`
			Success bool   `json:"success"`
		}
		var result Result

		err = x.MustUnmarshal(message, &result)
		if err != nil {
			ava.Errorf("host=%s |token=%s |err=%v", host, accessToken, err)
			return nil, err
		}

		if result.Type != "auth_ok" {
			ava.Errorf("host=%s |token=%s", host, accessToken)
			return nil, fmt.Errorf("authentication failed")
		}

		ava.Debugf("handshake success |host=%s |token=%s", host, accessToken)

		// 监听状态变化
		var state = struct {
			Id        int    `json:"id"`
			Type      string `json:"type"`
			EventType string `json:"event_type"`
		}{Id: time.Now().Nanosecond(), Type: "subscribe_events", EventType: "state_changed"}

		err = conn.WriteJSON(&state)
		if err != nil {
			ava.Errorf("host=%s |token=%s |err=%v", host, accessToken, err)
			return nil, err
		}

		_, stateMessage, err := conn.ReadMessage()
		if err != nil {
			ava.Errorf("host=%s |token=%s |err=%v", host, accessToken, err)
			return nil, err
		}

		ava.Debugf("state_changed |message=%s", string(stateMessage))

		var stateResult Result

		err = x.MustUnmarshal(stateMessage, &stateResult)
		if err != nil {
			ava.Errorf("host=%s |token=%s |err=%v", host, accessToken, err)
			return nil, err
		}

		if !stateResult.Success {
			ava.Errorf("host=%s |token=%s |stateResult=%v", host, accessToken, stateResult)
			return nil, fmt.Errorf("state subscription failed")
		}

		gHub.addConn(home, conn)
		return conn, nil
	}

	var backoffTime = time.Second * 10 // 初始重连时间

	for {
		conn, err := reconnect()
		if err != nil {
			ava.Errorf("initial connection failed, retrying in %v |err=%v", backoffTime, err)
			time.Sleep(backoffTime)
			backoffTime *= 2
			continue
		}

		backoffTime = time.Second * 10 // 连上之后，重置重连时间
		done := make(chan struct{})

		go func() {
			defer func() { close(done); conn.Close() }()

			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					ava.Errorf("home=%s |err=%v", home, err)
					return
				}

				c := ava.Background()
				c.Debug("----------设备状态变更---------", x.BytesToString(message))

				var fromState stateData
				err = x.MustUnmarshal(message, &fromState)
				if err != nil {
					c.Error(err)
					continue
				}

				if fromState.Type == "result" {
					continue
				}

				if fromState.Event.Data.NewState.State == "unavailable" {
					continue
				}

				var now = time.Now().UTC()

				//首先判断timestamp
				if v := fromState.Event.Data.NewState.Attributes.Timestamp; v != "" {
					// 使用 time.RFC3339Nano 解析日期字符串
					timestamp, err := time.Parse(time.RFC3339Nano, v)
					if err != nil {
						c.Error(err)
						continue
					}
					// 计算时间差
					duration := now.Sub(timestamp.UTC())

					// 判断时间差是否大于10分钟
					if duration > 10*time.Minute {
						continue
					}
				}

				//最近一次状态变化的时间如果跟更新时间不一致则跳过
				lastChanged, err := time.Parse(time.RFC3339Nano, fromState.Event.Data.NewState.LastChanged)
				if err != nil {
					c.Error(err)
					continue
				}

				lastUpdated, err := time.Parse(time.RFC3339Nano, fromState.Event.Data.NewState.LastUpdated)
				if err != nil {
					c.Error(err)
					continue
				}

				// 计算时间差
				duration := lastUpdated.Sub(lastChanged.UTC())

				// 判断时间差是否大于10分钟
				if duration > time.Minute {
					continue
				}

				//小米音响识别用户的语音数据
				if isXiaoMiSpeaker(fromState.Event.Data.NewState.EntityID) {
					err = receiveMessage(c, home, fromState.Event.Data.NewState.State, true)
					if err != nil {
						c.Error(err)
						continue
					}
				}

				//传感器数据
				if isHumanBodySensor(fromState.Event.Data.NewState.Attributes.DeviceClass) {
					var message string
					//人来
					if fromState.Event.Data.NewState.State == "on" {
						//message = "我是卧室人体传感器，人体传感器状态为on，请操作其它设备"
						message = "我是%s，人体传感器状态为on，请操作其它设备"
					}
					//人走
					if fromState.Event.Data.NewState.State == "off" {
						//message = "我是卧室人体传感器，人体传感器状态为off，请操作其它设备"
						message = "我是%s，人体传感器状态为off，请操作其它设备"
					}

					//err = receiveMessage(c, home, message)
					err = receiveMessage(c, home, fmt.Sprintf(message, fromState.Event.Data.NewState.Attributes.FriendlyName), false)
					if err != nil {
						c.Error(err)
						continue
					}
				}

			}
		}()

		select {
		case <-done:
			gHub.removeConn(home)
			break
		case <-interrupt:
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
