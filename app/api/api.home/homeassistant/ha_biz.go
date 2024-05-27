package homeassistant

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
	"vinesai/internel/ava"
	"vinesai/internel/x"
)

// 用户对应的url每个用户都有不同的端口号,后期配置成域名
// 手动后台配置,123是家庭标识符,后期做授权
// 授权之后才能使用家庭的api
var mapHome2Url = map[string]string{
	"123": "http://127.0.0.1:8123",
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
	"timer":                   true,
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
	req, _ := http.NewRequest("GET", mapHome2Url[home]+"/api/services", nil)
	req.Header.Set("Authorization", mapUserToken[home])
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.Error(err)
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	//c.Info(x.BytesToString(body))

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

func getStates(c *ava.Context, home string) (string, error) {
	req, _ := http.NewRequest("GET", mapHome2Url[home]+"/api/states", nil)
	req.Header.Set("Authorization", mapUserToken[home])
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.Error(err)
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	//c.Info(x.BytesToString(body))

	var filter = make([]map[string]interface{}, 0, 100)
	err = x.MustUnmarshal(body, &filter)
	if err != nil {
		c.Error(err)
		return "", err
	}

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

	return x.MustMarshal2String(result), nil

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

func callService(c *ava.Context, home, service string, data []byte) {
	c.Debugf("callService |url=%s |data=%s", mapHome2Url[home]+"/api/services/"+service, string(data))
	req, _ := http.NewRequest("POST", mapHome2Url[home]+"/api/services/"+service, bytes.NewReader(data))
	req.Header.Set("Authorization", mapUserToken[home])
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		c.Error(err)
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Error(err)
		return
	}

	c.Infof("callService |body=%s", x.BytesToString(body))
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
{"command":[{"data":{"entity_id":"light.smart_led_strip_2","rgb_color":[255,255,0]},"service":"light/turn_on"}],"message":"好的主人，已将客厅led改为黄色"}
2.关闭【客厅led】:
{"command":[{"data":{"entity_id":"light.smart_led_strip_2"},"service":"light/turn_off"}],"message":"客厅led已乖乖睡觉啦~"}
3.打开【客厅插座】:
{"command":[{"data":{"entity_id":"switch.qmi_psv3_4067_switch"},"service":"switch/turn_on"}],"message":"客厅插座已打开"}
请直接给出 JSON 格式的指令结果,不要有其他文字。当前的指令清单和设备清单如下: 【指令清单】:%s 【设备清单】:%s`

var aiTmp2 = `你是一个智能家居助手,负责理解用户意图，然后通过 Home Assistant REST API 控制家中的智能设备。
你能控制的设备及其指令如下：

【设备清单】:%s，【指令清单】:%s 。

请根据用户意图，选择合适的设备和指令执行相应的操作,并以 JSON 格式返回指令结果。
例如:
1.将【客厅led】改为黄色:
{"command":[{"data":{"entity_id":"light.smart_led_strip_2","rgb_color":[255,255,0]},"service":"light/turn_on"}],"message":"好的主人，已将客厅led改为黄色"}
2.关闭【客厅led】:
{"command":[{"data":{"entity_id":"light.smart_led_strip_2"},"service":"light/turn_off"}],"message":"客厅led已乖乖睡觉啦~"}
3.打开【客厅插座】:
{"command":[{"data":{"entity_id":"switch.qmi_psv3_4067_switch"},"service":"switch/turn_on"}],"message":"客厅插座已打开"}
4.设备不可用:
{"command":[{"data":{"entity_id":"switch.smart_plug_socket_1","state":"unavailable"},"service":""}],"message":"卧室插座不可用"}

注意：
在判断设备位置时,请使用 friendly_name 字段。指令结果中包含以下字段:
- service:结合指令清单中的 【domain】 和 【services】 得到,如 "light/turn_on" 。
- data:根据设备当前状态和指令要求,生成修改后的设备数据 。
- message:以友好、俏皮的口吻告知操作结果。`

//- 【设备清单】中如果是{ "entity_id": "switch.smart_plug_socket_1", "state": "unavailable"},则提示我“xxx不可用”,参考第4点。
