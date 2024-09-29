package tuya

//
//import (
//	"context"
//	"fmt"
//	"net/http"
//	"reflect"
//	"strings"
//	"sync/atomic"
//	"time"
//	"vinesai/internel/ava"
//	"vinesai/internel/lib/connector"
//	"vinesai/internel/x"
//	"vinesai/proto/ptuya"
//
//	"github.com/panjf2000/ants/v2"
//)
//
//type Tuya struct {
//}
//
//// 查询用户家庭信息
//func (t *Tuya) HomeList(c *ava.Context, req *ptuya.HomeListReq, rsp *ptuya.HomeListRsp) {
//	var resp = &struct {
//		Result []*ptuya.HomeListData `json:"result"`
//	}{}
//
//	err := connector.MakeGetRequest(
//		context.Background(),
//		connector.WithAPIUri(fmt.Sprintf("/v1.0/users/%s/homes", defaultUid)),
//		connector.WithResp(resp),
//	)
//
//	if err != nil {
//		c.Error(err)
//		rsp.Code = http.StatusInternalServerError
//		rsp.Msg = "获取用户家庭信息失败"
//		return
//	}
//
//	rsp.Code = http.StatusOK
//	rsp.Data = resp.Result
//}
//
//func (t *Tuya) Code(c *ava.Context, req *ptuya.LoginCodeReq, rsp *ptuya.LoginCodeRsp) {
//	//todo 校验手机号规则是否正确
//	//todo 验证码发送流程
//	rsp.Code = http.StatusOK
//	rsp.Data = &ptuya.LoginCodeData{
//		Code: "123456",
//	}
//}
//
//var defaultSchema = "homingai"
//
//// 用户登录，获取jwttoken和涂鸦用户的uid
//// ome app或者涂鸦sdk开发用到
//func (t *Tuya) Login(c *ava.Context, req *ptuya.LoginReq, rsp *ptuya.LoginRsp) {
//	if req.Phone == "" {
//		rsp.Code = http.StatusBadRequest
//		rsp.Msg = "手机号不能为空"
//		return
//	}
//
//	if req.Password == "" {
//		rsp.Code = http.StatusBadRequest
//		rsp.Msg = "密码不能为空"
//		return
//	}
//
//	type r struct {
//		Result struct {
//			Uid string `json:"uid"`
//		} `json:"result"`
//		Success bool `json:"success"`
//	}
//
//	type Payload struct {
//		CountryCode  string `json:"country_code"`
//		Username     string `json:"username"`
//		Password     string `json:"password"`
//		UsernameType int    `json:"username_type"`
//		NickName     string `json:"nick_name"`
//		TimeZoneID   string `json:"time_zone_id"`
//	}
//
//	var resp = &r{}
//	var payload = &Payload{
//		CountryCode:  "86",
//		Username:     req.Phone,
//		Password:     x.Md5(req.Password),
//		UsernameType: 1,
//		TimeZoneID:   "Asia/Shanghai",
//	}
//
//	//请求涂鸦获取用户的uid
//	//https://developer.tuya.com/cn/docs/cloud/76f3e0885f?id=Kawfji9n0sdmq#title-1-%E6%8E%A5%E5%8F%A3%E5%9C%B0%E5%9D%80
//	//通过获取用户列表输入用户名86-手机号，可以查询用户id
//	err := connector.MakePostRequest(
//		context.Background(),
//		connector.WithAPIUri(fmt.Sprintf("/v1.0/apps/%s/user", defaultSchema)),
//		connector.WithPayload(x.MustMarshal(payload)),
//		connector.WithResp(resp),
//	)
//
//	if err != nil {
//		c.Error(err)
//		rsp.Code = http.StatusInternalServerError
//		rsp.Msg = "用户信息同步失败"
//		return
//	}
//
//	if resp.Result.Uid == "" {
//		c.Debugf("resp.Result.Uid is empty |resp=%s", x.MustMarshal2String(resp))
//		rsp.Code = http.StatusInternalServerError
//		rsp.Msg = "用户信息为空"
//		return
//	}
//
//	//组装jwt token
//	token := generateJWToken(c, req.Phone, resp.Result.Uid)
//
//	rsp.Code = http.StatusOK
//	rsp.Data = &ptuya.LoginData{Token: token}
//}
//
//// todo 暂时写死用户id
//var defaultUid = "ay1716438065043jAiE1"
//
//// 用户意图，通过设备列表和控制指令集发送给ai，获取到返回的播报和指令，并发送指令到涂鸦
//func (t *Tuya) Intent(c *ava.Context, req *ptuya.UserIntentReq, rsp *ptuya.UserIntentRsp) {
//	if req.Content == "" {
//		rsp.Code = http.StatusBadRequest
//		rsp.Msg = "content不能为空"
//		return
//	}
//
//	if req.HomeId == 0 {
//		rsp.Code = http.StatusBadRequest
//		rsp.Msg = "home_id不能为空"
//		return
//	}
//
//	var now = time.Now()
//	//获取设备列表
//	var resp = &deviceListResp{}
//	err := connector.MakeGetRequest(
//		context.Background(),
//		connector.WithAPIUri(fmt.Sprintf("/v1.0/homes/%d/devices", req.HomeId)),
//		connector.WithResp(resp),
//	)
//
//	if err != nil {
//		c.Error(err)
//		rsp.Code = http.StatusInternalServerError
//		rsp.Msg = "获取用户设备列表失败"
//		return
//	}
//
//	//去掉离线设备
//	var onlineDevice = make([]*device, 0, 30)
//	for i := range resp.Result {
//		if resp.Result[i].Online {
//			onlineDevice = append(onlineDevice, resp.Result[i])
//		}
//	}
//
//	if len(onlineDevice) == 0 {
//		rsp.Code = http.StatusInternalServerError
//		rsp.Msg = "当前用户没有设备"
//		return
//	}
//
//	c.Debugf("-------------------获取设备列表---------------%v", time.Now().Sub(now).Seconds())
//
//	var now_11 = time.Now()
//	//通过ai获取需要控制的设备
//	shortDeviceList, fullDevices, ids,err := deviceListGpt(c, req.Content, onlineDevice)
//	//shortDeviceList, fullDevices, ids, err := deviceListQianwen(c, req.Content, onlineDevice)
//	if err != nil {
//		c.Error(err)
//		rsp.Code = http.StatusInternalServerError
//		rsp.Msg = "服务器内部错误"
//		return
//	}
//
//	c.Debugf("-------------------通过ai获取需要控制的设备---------------%v", time.Now().Sub(now_11).Seconds())
//
//	if len(shortDeviceList) == 0 {
//		rsp.Code = http.StatusInternalServerError
//		rsp.Msg = "没有需要控制的设备"
//		return
//	}
//
//	var cmd []*command
//	for i := 0; i < len(ids); i += 20 {
//		// 计算当前批次的结束索引，确保不超过数组长度
//		end := i + 20
//		if end > len(ids) {
//			end = len(ids)
//		}
//
//		var fs = &commandsResp{}
//
//		//获取用户设备指令,一次最多20个设备
//		err = connector.MakeGetRequest(
//			context.Background(),
//			connector.WithAPIUri(fmt.Sprintf("/v1.0/devices/functions?device_ids=%s", strings.Join(ids[i:end], ","))),
//			connector.WithResp(fs),
//		)
//
//		if err != nil {
//			c.Error(err)
//			rsp.Code = http.StatusInternalServerError
//			rsp.Msg = "获取指令失败"
//			return
//		}
//
//		cmd = append(cmd, fs.Result...)
//	}
//
//	var now_1 = time.Now()
//	//请求ai获取控制指令和播报消息
//	msg2AiResp, err := msg2Gpt(c, req.Content, x.MustMarshal2String(cmd), x.MustMarshal2String(shortDeviceList))
//	//msg2AiResp, err := msg2Qianwen(c, req.Content, x.MustMarshal2String(cmd), x.MustMarshal2String(shortDeviceList))
//	if err != nil {
//		c.Error(err)
//		rsp.Code = http.StatusInternalServerError
//		rsp.Msg = "内部请求错误"
//		return
//	}
//
//	c.Debugf("------------------请求ai获取控制指令和播报消息---------------%v", time.Now().Sub(now_1).Seconds())
//
//	type controlDeviceResp struct {
//		Result  bool   `json:"result"`
//		Success bool   `json:"success"`
//		T       int    `json:"t"`
//		Tid     string `json:"tid"`
//	}
//
//	var commandCount int64 = 0
//
//	pool, err := ants.NewPool(len(msg2AiResp.Result))
//	if err != nil {
//		c.Error(err)
//		rsp.Code = http.StatusInternalServerError
//		rsp.Msg = "内部请求错误"
//		return
//	}
//
//	//并发发起设备控制
//	for i := range msg2AiResp.Result {
//		var tmpResp = msg2AiResp.Result[i]
//
//		_ = pool.Submit(func() {
//			var cdResp = &controlDeviceResp{}
//			var isZero = false
//			//再次筛选设备，除去状态一致的设备
//			if devices, devicesIsExist := fullDevices[tmpResp.Id]; devicesIsExist {
//				for iii2 := range tmpResp.Data.Commands {
//					var s = tmpResp.Data.Commands[iii2]
//
//					for iii := range devices.Status {
//
//						var data = devices.Status[iii]
//						if data.Code != s.Code {
//							continue
//						}
//
//						var isEqual = false
//						//判断数据是不是string,且是不是json的情况
//						if cmdValue, cmdIsString := s.Value.(string); cmdIsString {
//							//判断命令状态值是不是json
//							var tMap = make(map[string]interface{})
//							err = x.MustUnmarshal([]byte(cmdValue), &tMap)
//							if err == nil {
//								//判断设备状态值是不是string
//								var deviceValue, deviceIsString = data.Value.(string)
//								if !deviceIsString {
//									continue
//								}
//								//判断设备状态值是不是json
//								var bMap = make(map[string]interface{})
//								err = x.MustUnmarshal([]byte(deviceValue), &bMap)
//								if err != nil {
//									continue
//								}
//
//								//比较设备状态值和命令状态值是否相等
//								if reflect.DeepEqual(bMap, tMap) {
//									isEqual = true
//								}
//							}
//						} else {
//
//							//比较设备状态值和命令状态值是否相等
//							if reflect.DeepEqual(s.Value, data.Value) {
//								isEqual = true
//
//							}
//						}
//
//						if isEqual {
//							//如果设备状态值和命令状态值相等，删除命令
//							//如果只有一条数据直接返回
//							if len(tmpResp.Data.Commands) == 1 {
//								tmpResp.Data.Commands = tmpResp.Data.Commands[0:0]
//								isZero = true
//								break
//							}
//							tmpResp.Data.Commands = append(tmpResp.Data.Commands[:iii], tmpResp.Data.Commands[iii+1:]...)
//						}
//					}
//				}
//			}
//
//			if isZero {
//				return
//			}
//
//			err = connector.MakePostRequest(
//				context.Background(),
//				connector.WithAPIUri(fmt.Sprintf("/v1.0/devices/%s/commands", tmpResp.Id)),
//				connector.WithPayload(x.MustMarshal(tmpResp.Data)),
//				connector.WithResp(cdResp),
//			)
//
//			if err != nil {
//				c.Error(err)
//				return
//			}
//
//			if cdResp.Result && cdResp.Success {
//				atomic.AddInt64(&commandCount, 1)
//			}
//		})
//	}
//
//	err = pool.ReleaseTimeout(time.Second * 10)
//	if err != nil {
//		c.Error(err)
//		rsp.Code = http.StatusInternalServerError
//		rsp.Msg = "内部请求错误"
//	}
//
//	if commandCount == 0 {
//		rsp.Code = http.StatusOK
//		rsp.Msg = "设备已经是你想要的状态了，不需要控制"
//		return
//	}
//
//	rsp.Code = http.StatusOK
//	rsp.Data = &ptuya.UserIntentData{Content: msg2AiResp.Voice}
//}
