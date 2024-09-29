package tuya

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
	"vinesai/internel/ava"
	"vinesai/internel/db"
	"vinesai/internel/lib/tuyago"
	"vinesai/internel/x"
	"vinesai/proto/phub"

	"github.com/gogo/protobuf/proto"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/panjf2000/ants/v2"
)

var jwtKey = []byte("AOGQ6MNVIU9Y5J5LK0PWB1A8H2Z4ERCB")

type tokenClaims struct {
	Timestamp string
	Uid       string
	Phone     string
	jwtv5.RegisteredClaims
}

func parseJWToken(token string) (*tokenClaims, error) {
	t, err := jwtv5.ParseWithClaims(token, &tokenClaims{}, func(t *jwtv5.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := t.Claims.(*tokenClaims); ok && t.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

var defaultExpiryDelta = time.Hour * 24 * 30 * 12 * 100

func generateJWToken(c *ava.Context, phone, uid string) string {
	expiry := jwtv5.NewNumericDate(x.LocalTimeNow().Add(defaultExpiryDelta))
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, tokenClaims{
		Timestamp: x.LocalTimeNow().Format(time.RFC3339),
		Uid:       uid,
		Phone:     phone,
		RegisteredClaims: jwtv5.RegisteredClaims{
			Issuer:    "homingai",
			Subject:   uid,
			Audience:  []string{"homingai"},
			ExpiresAt: expiry,
			NotBefore: jwtv5.NewNumericDate(x.LocalTimeNow()), //token在此时间之前不能被接收处理
			IssuedAt:  jwtv5.NewNumericDate(x.LocalTimeNow()),
			ID:        uuid.New().String(),
		},
	})

	str, err := token.SignedString(jwtKey)
	if err != nil {
		c.Errorf("generateJWToken |err=%v", err)
		return ""
	}

	return str
}

var whiteHttpPathList = map[string]bool{
	"/home/tuya/login": true,
}

func Authorization(c *ava.Context) (proto.Message, error) {

	if _, ok := whiteHttpPathList[c.Metadata.Method()]; ok {
		return nil, nil
	}

	var rsp phub.CommonData
	//处理bearar
	token := c.GetHeader("Authorization")
	if token == "" {
		rsp.Code = 401
		rsp.Msg = "请求头Authorization信息缺失"
		return &rsp, errors.New("请求头Authorization信息缺失")
	}

	t, err := parseJWToken(token)
	if err != nil {
		rsp.Code = 401
		rsp.Msg = "身份认证失败"
		return &rsp, errors.New("身份认证失败")
	}

	c.Infof("Oauth |data=%v", x.MustMarshal2String(t))

	c.Set("X-Tuya-Uid", t.Uid)
	return nil, nil
}

func getTuyaUid(c *ava.Context) string {
	return c.GetString("X-Tuya-Uid")
}

// ai返回内容格式
type aiResp struct {
	Voice  string       `json:"voice"`
	Result []aiRespData `json:"result"`
}

type aiRespData struct {
	Id   string `json:"id"`
	Data struct {
		Commands []status `json:"commands"`
	} `json:"data"`
}

// 获取用户的设备列表
type deviceListResp struct {
	Result   []*device `json:"result"`
	Position string    `json:"position"` //非接口返回字段，需要通过其他接口获取
	Success  bool      `json:"success"`
	T        int       `json:"t"`
	Tid      string    `json:"tid"`
}

type shortDeviceListResp struct {
	Devices  []*shortDevice `json:"devices"`
	Position string         `json:"position"` //非接口返回字段，需要通过其他接口获取
}

type function struct {
	Code   string      `json:"code"`
	Desc   string      `json:"desc"`
	Name   string      `json:"name"`
	Type   string      `json:"type"`
	Values interface{} `json:"values"`
}

type command struct {
	Devices   []string    `json:"devices"`
	Functions []*function `json:"functions"`
}

// 批量获取指令集
type commandsResp struct {
	Result  []*command `json:"result"`
	Success bool       `json:"success"`
	T       int        `json:"t"`
	Tid     string     `json:"tid"`
}

type status struct {
	Code  string      `json:"code"`
	Value interface{} `json:"value"`
}

// 用户设备信息
type device struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	Status    []*status `json:"status"`
	Category  string    `json:"category"`
	Online    bool      `json:"online"`
	ProductId string    `json:"product_id"`
}

type shortDevice struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

// 根据ai返回的简短数据找出全量数据
func shortDeviceInfo2Devices(s []*shortDevice, l []*device) (map[string]*device, []string) {
	// 创建一个 map，键是设备 ID，值是完整的 device 结构体
	deviceMap := make(map[string]*device)

	// 将完整的设备信息存储到 map 中
	for _, d := range l {
		deviceMap[d.Id] = d
	}

	// 准备一个切片保存匹配的设备
	var result = make(map[string]*device, len(s))

	var ids = make([]string, 0, len(s))
	// 根据 shortDeviceInfo 中的 ID 找出对应的完整设备信息
	for _, shortInfo := range s {
		if fullDevice, exists := deviceMap[shortInfo.Id]; exists {
			result[shortInfo.Id] = fullDevice
			ids = append(ids, shortInfo.Id)
		}
	}

	return result, ids
}

type room struct {
	Name   string `json:"name"`
	RoomID int64  `json:"room_id"`
}

type roomInfo struct {
	Result struct {
		GeoName string  `json:"geo_name"`
		HomeID  int64   `json:"home_id"`
		Lat     float64 `json:"lat"`
		Lon     float64 `json:"lon"`
		Name    string  `json:"name"`
		Rooms   []*room `json:"rooms"`
	} `json:"result"`
	Success bool  `json:"success"`
	T       int64 `json:"t"`
}

// 设备筛选获取设备列表,发送设备数据和指令数据给ai，让ai根据用户意图，选择控制设备，并返回设备指令
func invokeAI(c *ava.Context, homeId int32, content string) (*aiResp, map[string]*device, error) {
	//获取房间信息
	var roomResp = &roomInfo{}

	err := tuyago.Get(c, fmt.Sprintf("/v1.0/homes/%d/rooms", homeId), roomResp)

	if err != nil {
		c.Error(err)
		return nil, nil, err
	}

	//判断房间是否在用户的意图中
	var r = make([]*room, 0, 4)
	var tmpR = make([]*room, 0, len(roomResp.Result.Rooms))
	for i := range roomResp.Result.Rooms {
		tmpR = append(tmpR, roomResp.Result.Rooms[i])
		if strings.Contains(content, roomResp.Result.Rooms[i].Name) {
			r = append(r, roomResp.Result.Rooms[i])
			continue
		}
	}

	if len(r) == 0 {
		r = nil
		r = tmpR
	}

	var mux = new(sync.Mutex)
	var deviceShortList []*shortDeviceListResp
	var deviceList []*deviceListResp
	var productIdMap = make(map[string]string) //对应设备id,当查从redis查询不到指令时，用设备id从涂鸦获取
	var productId []string
	var deviceFullInfoMap = make(map[string]*device, 10)

	pool, err := ants.NewPool(len(roomResp.Result.Rooms))
	if err != nil {
		c.Error(err)
		return nil, nil, err
	}

	//遍历房间获取设备
	for i := range r {
		var tmpRoom = r[i]
		err = pool.Submit(func() {
			var tmpResp = &deviceListResp{}

			err := tuyago.Get(c, fmt.Sprintf("/v1.0/homes/%d/rooms/%d/devices", homeId, tmpRoom.RoomID), tmpResp)

			if err != nil {
				ava.Error(err)
				return
			}

			if !tmpResp.Success {
				ava.Debugf("get device list from room fail |data=%v |id=%v", tmpResp, tmpRoom.RoomID)
				return
			}

			if len(tmpResp.Result) == 0 {
				return
			}

			//去掉离线设备
			var tmpShortDevice []*shortDevice
			var tmpDevice []*device

			for ii := range tmpResp.Result {
				tmpData := tmpResp.Result[ii]

				if tmpData.Online {

					tmpShortDevice = append(tmpShortDevice, &shortDevice{
						Id:   tmpData.Id,
						Name: tmpData.Name,
					})
					tmpDevice = append(tmpDevice, tmpData)

					mux.Lock()
					if _, ok := productIdMap[tmpData.ProductId]; !ok {
						productId = append(productId, tmpData.ProductId)
						productIdMap[tmpData.ProductId] = tmpData.Id
					}
					deviceFullInfoMap[tmpData.Id] = tmpData
					mux.Unlock()
				}
			}

			if len(tmpShortDevice) == 0 {
				return
			}

			mux.Lock()
			deviceList = append(deviceList, &deviceListResp{Result: tmpDevice, Position: tmpRoom.Name})
			deviceShortList = append(deviceShortList, &shortDeviceListResp{Devices: tmpShortDevice, Position: tmpRoom.Name})
			mux.Unlock()
		})
	}

	err = pool.ReleaseTimeout(time.Second * 10)
	if err != nil {
		c.Error(err)
		return nil, nil, err
	}

	if len(deviceShortList) == 0 {
		return nil, nil, errors.New("没有设备需要控制")
	}

	//从redis中获取设备指令，当指令不存在时从涂鸦api去获取设备指令
	values, err := db.GRedis.MGet(context.Background(), productId...).Result()
	if err != nil {
		c.Error(err)
		return nil, nil, err
	}

	var cmdsMap = make(map[string][]*function)
	for i, value := range values {

		var fs []*function

		if value == nil {

			var cmdResp = &commandsResp{}
			//从涂鸦api查询指令
			err := tuyago.Get(c, fmt.Sprintf("/v1.0/devices/functions?device_ids=%s", productId[i]), cmdResp)

			if err != nil {
				c.Error(err)
				continue
			}

			//将指令存到redis
			err = db.GRedis.Set(context.Background(), productId[i], x.MustMarshal2String(fs), 0).Err()
			if err != nil {
				c.Error(err)
				continue
			}

			if len(cmdResp.Result) == 0 {
				c.Debug("not found data")
				continue
			}

			fs = cmdResp.Result[0].Functions

		} else {
			err = x.MustUnmarshal([]byte(value.(string)), fs)
			if err != nil {
				c.Error(err)
				continue
			}
		}

		cmdsMap[productId[i]] = fs
	}

	out, err := langchainRun(c, defaultUid, content, x.MustMarshal2String(deviceShortList), x.MustMarshal2String(cmdsMap))
	if err != nil {
		c.Error(err)
		return nil, nil, err
	}

	var resultFullInfoDevice = make(map[string]*device)
	//根据需要控制的设备信息，找到全量设备信息
	for i := range out.Result {
		resultFullInfoDevice[out.Result[i].Id] = deviceFullInfoMap[out.Result[i].Id]
	}

	return out, resultFullInfoDevice, nil
}
