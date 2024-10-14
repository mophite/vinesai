package langchain

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
	"vinesai/internel/ava"
	"vinesai/internel/db"
	"vinesai/internel/lib/tuyago"
	"vinesai/internel/x"

	"github.com/redis/go-redis/v9"
	"github.com/tmc/langchaingo/llms"
)

// 一次多个意图
type summaries struct {
	Result []*summaryData `json:"result"`
}

type summaryData struct {
	Content string   `json:"content"`
	Summary string   `json:"summary"`
	Devices []string `json:"devices"`
}

type summary struct{ CallbacksHandler LogHandler }

func (s *summary) Name() string {
	return "summary"
}

func (s *summary) Description() string {
	return `明确的智能家居设备控制请求`
}

var defaultSummaryMsg = "x-langchaingo-summary-msg"

func setSummaryMsg(c *ava.Context, value string) {
	c.Set(defaultSummaryMsg, value)
}

func getSummaryMsg(c *ava.Context) (string, error) {
	s := c.GetString(defaultSummaryMsg)
	if len(s) == 0 {
		return "", errors.New("获取数据失败")
	}

	return s, nil
}

func (s *summary) Call(ctx context.Context, input string) (string, error) {

	var c = fromCtx(ctx)
	var homeId = getHomeId(c)

	//获取所有设备
	devicesName, devicesNameMap, err := getSummaryDevices(c, homeId)
	if err != nil {
		c.Error(err)
		return "", err
	}

	summary, err := getSummaryInfo(c, getFirstInput(c), devicesName)
	if err != nil {
		c.Error(err)
		return "", err
	}

	msg, err := chooseAndControlDevices(c, summary, devicesNameMap)
	if err != nil {
		c.Error(err)
		return "", err
	}

	c.Debug("-------msg----", msg)

	//var choice = []*llms.ContentChoice{
	//	{
	//		Content:    msg,
	//		StopReason: "stop",
	//		FuncCall:   nil,
	//		ToolCalls:  nil,
	//	},
	//}
	//
	//s.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, &llms.ContentResponse{Choices: choice})

	setSummaryMsg(c, msg)

	return msg, errors.New("exit")
}

type ShortSummaryDeviceInfo struct {
	Result []*ShortSummryDevice `json:"result"`
}

type ShortSummryDevice struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Position  string `json:"position"`
	ProductId string `json:"product_id"`
	Category  string `json:"category"`
}

// 根据位置信息选取设备
func chooseAndControlDevices(c *ava.Context, s *summaries, devicesMap map[string]*device) (string, error) {

	var failureMessage []string
	var successMessage []string
	var offlineMessage []string
	var alreadyMessage []string

	for i := range s.Result {
		summa := s.Result[i]
		for ii := range summa.Devices {
			name := summa.Devices[ii]
			d, ok := devicesMap[name]
			if !ok {
				continue
			}

			//从redis获取简短的设备指令
			fs, err := getSummaryCommand(c, d.Category, d.ProductId, summa.Summary, d.Id, d.Name, summa.Content)
			if err != nil {
				c.Error(err)
				continue
			}

			var tmpDevicesResp = &deviceResp{}

			err = tuyago.Get(c, fmt.Sprintf("/v1.0/devices/%s", d.Id), tmpDevicesResp)

			if err != nil {
				c.Error(err)
				continue
			}

			if !tmpDevicesResp.Success {
				c.Debugf("get device list from room fail |data=%v |id=%v", tmpDevicesResp, d.Id)
				continue
			}

			//判断设备状态是否在线
			if !tmpDevicesResp.Result.Online {
				offlineMessage = append(offlineMessage, offlineMsg(d.Name))
				continue
			}

			var isSame = false
			//判断设备当前状态是否和指令返回一致
			for iii := range fs.Result {
				tmpFs := fs.Result[iii]
				for iiii := range tmpDevicesResp.Result.Status {
					tmpDeviceStatus := tmpDevicesResp.Result.Status[iiii]

					if tmpFs.Code == tmpDeviceStatus.Code {

						if reflect.DeepEqual(tmpFs.Value, tmpDeviceStatus.Value) {
							isSame = true
						} else {
							isSame = false
						}

						break
					}
				}
			}

			//状态一致不用比较
			if isSame {
				alreadyMessage = append(alreadyMessage, alreadyMsg(d.Name))
				continue
			}

			//执行设备控制
			var controlResp summaryControlDeviceResp
			//执行控制
			err = tuyago.Post(c, fmt.Sprintf("/v1.0/devices/%s/commands", d.Id), &summaryControlData{Commands: x.MustMarshal2String(fs.Result)}, &controlResp)

			if err != nil {
				c.Error(err)
				failureMessage = append(failureMessage, failureMsg(name))
				continue
			}

			if controlResp.Result && controlResp.Success {
				//判断语气中是否包含设备名称，这种情况是通过ai获取的结果
				if strings.Contains(fs.SuccessMsg, d.Name) {
					successMessage = append(successMessage, fs.SuccessMsg)
					fs.SuccessMsg = strings.Trim(fs.SuccessMsg, d.Name)
				} else {
					successMessage = append(successMessage, d.Name+fs.SuccessMsg)
				}

				//缓存指令
				err = db.GRedis.HSet(
					context.Background(),
					getSummaryCategoryListKey(d.Category, d.ProductId),
					summa.Summary,
					x.MustMarshal2String(fs)).Err()
				if err != nil {
					c.Error(err)
				}
			} else {
				failureMessage = append(failureMessage, failureMsg(name))
			}
		}
	}

	var msg string

	if len(failureMessage) > 3 {
		msg = "有大量设备控制失败，请检查"
		return msg, nil
	}

	if len(successMessage) > 3 || len(alreadyMessage) > 3 {
		msg = "好的主人，设备都已控制成功啦"
		return msg, nil
	}

	if len(offlineMessage) > 3 {
		msg = "有大量设备离线，请检查"
		return msg, nil
	}

	for i := range failureMessage {
		msg += failureMessage[i] + ","
	}

	for i := range successMessage {
		msg += successMessage[i] + ","
	}

	for i := range offlineMessage {
		msg += offlineMessage[i] + ","
	}

	for i := range alreadyMessage {
		msg += alreadyMessage[i] + ","
	}

	return msg, nil
}

// 设备，动作，值
func successMsg(name, successMsg string) string {
	return name + successMsg
}

func failureMsg(name string) string {
	return name + "控制失败"
}

func offlineMsg(name string) string {
	return name + "已离线"
}

func alreadyMsg(name string) string {
	return name + "已经是你想要的状态了"
}

type summaryControlDeviceResp struct {
	Result  bool   `json:"result"`
	Success bool   `json:"success"`
	T       int    `json:"t"`
	Tid     string `json:"tid"`
}

type summaryControlData struct {
	Commands string `json:"commands"`
}

var defaultSummaryCommandKey = "TUYA_SUMMARY_COMMAND_%s_%s"
var defaultSummaryNativeCommandKey = "TUYA_SUMMARY_NATIVE_COMMAND_"

func getSummaryCategoryListKey(category, productId string) string {
	return fmt.Sprintf(defaultSummaryCommandKey, category, productId)
}

func getProductIdCommand(c *ava.Context, productId, deviceId string) (string, error) {

	//先从redis获取设备完整指令
	result, err := db.GRedis.Get(context.Background(), defaultSummaryNativeCommandKey+productId).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		c.Error(err)
		return "", err
	}

	//有些设备没有指令
	if result == "-" {
		return "-", fmt.Errorf("this productID=%s no function", productId)
	}

	if result != "" {
		return result, nil
	}

	//从涂鸦获取设备指令
	var cmdResp = &commandsResp{}
	//从涂鸦api查询指令
	err = tuyago.Get(c, fmt.Sprintf("/v1.0/devices/functions?device_ids=%s", deviceId), cmdResp)

	if err != nil {
		c.Errorf("productId=%s |err=%v", productId, err)
		return "", err
	}

	if !cmdResp.Success {
		return "", fmt.Errorf("get productId=%s command failure", productId)
	}

	var value = "-"
	if cmdResp.Success && len(cmdResp.Result) != 0 {
		value = x.MustMarshal2String(cmdResp.Result[0].Functions)
	}

	err = db.GRedis.Set(context.Background(), defaultSummaryNativeCommandKey+productId, value, 0).Err()
	if err != nil {
		ava.Error(err)
		return "", err
	}

	if value == "-" {
		return "-", fmt.Errorf("this productID=%s no function", productId)
	}

	return value, nil
}

type summaryCommandsResp struct {
	SuccessMsg string `json:"success_msg"`
	Result     []struct {
		Code  string      `json:"code"`
		Value interface{} `json:"value"`
	} `json:"result"`
}

func getSummaryCommand(c *ava.Context, category, productId, summaryStr, deviceId, deviceName, input string) (*summaryCommandsResp, error) {

	var key = getSummaryCategoryListKey(category, productId)
	var resp summaryCommandsResp

	//先从redis获取设备指令
	result, err := db.GRedis.HGet(context.Background(), key, summaryStr).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		c.Error(err)
		return nil, err
	}

	//有些设备没有指令
	if result == "-" {
		return nil, fmt.Errorf("this productID=%s category=%s no function", category, productId)
	}

	if result != "" {
		err = x.MustNativeUnmarshal([]byte(result), &resp)
		if err != nil {
			c.Error(err)
			return nil, err
		}
		return &resp, nil
	}

	//获取完整指令
	pResult, err := getProductIdCommand(c, productId, deviceId)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	//通过ai获取指令数据
	mcList := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(fmt.Sprintf(summaryCommandPrompts, deviceName, pResult))},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart(input)},
		},
	}

	err = generateContentWithout(c, mcList, &resp)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	if len(resp.Result) == 0 {
		return nil, errors.New("no command response")
	}

	return &resp, nil
}

// 同步用户设备信息
// 设备名称数组
// 设备名：详细设备数据
func syncDevicesForSummary(c *ava.Context, homeId string) ([]string, map[string]*device, error) {

	//获取房间信息
	var roomResp = &roomInfo{}

	err := tuyago.Get(c, fmt.Sprintf("/v1.0/homes/%s/rooms", homeId), roomResp)

	if err != nil {
		c.Error(err)
		return nil, nil, err
	}

	if !roomResp.Success {
		return nil, nil, errors.New("获取房间信息失败")
	}

	c.Debugf("roomInfo: %v", roomResp)

	if len(roomResp.Result.Rooms) == 0 {
		return nil, nil, errors.New("请创建房间，并将设备添加到房间中")
	}

	var devicesName = make([]string, 0, 20)
	var devicesNameMap = make(map[string]*device, 20)

	//遍历房间获取设备
	for i := range roomResp.Result.Rooms {
		var tmpRoom = roomResp.Result.Rooms[i]

		var tmpDevicesResp = &deviceListResp{}

		err = tuyago.Get(c, fmt.Sprintf("/v1.0/homes/%s/rooms/%d/devices", homeId, tmpRoom.RoomID), tmpDevicesResp)

		if err != nil {
			ava.Error(err)
			return nil, nil, err
		}

		if !tmpDevicesResp.Success {
			ava.Debugf("get device list from room fail |data=%v |id=%v", tmpDevicesResp, tmpRoom.RoomID)
			continue
		}

		c.Debugf("所有设备 ｜homeId=%s |rooName=%s |tmpResp=%v", homeId, tmpRoom.Name, x.MustMarshal2String(tmpDevicesResp))

		if len(tmpDevicesResp.Result) == 0 {
			continue
		}

		for ii := range tmpDevicesResp.Result {
			tmpDeviceData := tmpDevicesResp.Result[ii]

			//如果设备品类不在控制范围内，则不添
			if getCategoryName(tmpDeviceData.Category) == "" {
				continue
			}

			//判断设备名称中是否包含房间位置
			if !strings.Contains(tmpDeviceData.Name, tmpRoom.Name) {
				//修改设备名称
				renameBody := &struct {
					Name string `json:"name"`
				}{
					Name: tmpRoom.Name + tmpDeviceData.Name,
				}
				renameResp := &struct {
					Result bool `json:"result"`
				}{}
				err = tuyago.Put(c, "/v1.0/devices/"+tmpDeviceData.Id, renameBody, renameResp)
				if err != nil {
					c.Error(err)
					break
				}

				if !renameResp.Result {
					break
				}

				tmpDeviceData.Name = tmpRoom.Name + tmpDeviceData.Name

			}
			devicesName = append(devicesName, tmpDeviceData.Name)
			devicesNameMap[tmpDeviceData.Name] = tmpDeviceData
		}
	}

	return devicesName, devicesNameMap, nil
}

// 单个房间里的设备类型
type SummaryCategories struct {
	CategoryData []*summaryCategoryData `json:"category_data"`
}

type summaryCategoryData struct {
	CategoryName string `json:"category_name"`
	Category     string `json:"category"`
}

var redisKeyTuyaSummaryDeviceName = "TUYA_SUMMARY_DEVICE_NAME_"
var redisKeyTuyaSummaryDeviceNameMap = "TUYA_SUMMARY_DEVICE_NAME_MAP_"

func getSummaryDevices(c *ava.Context, homeId string) ([]string, map[string]*device, error) {
	devicesNameResult, err := db.GRedis.Get(context.Background(), redisKeyTuyaSummaryDeviceName+homeId).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		c.Error(err)
		return nil, nil, err
	}

	deviceNameMapResult, err := db.GRedis.Get(context.Background(), redisKeyTuyaSummaryDeviceNameMap+homeId).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		c.Error(err)
		return nil, nil, err
	}

	if devicesNameResult == "" || errors.Is(err, redis.Nil) || deviceNameMapResult == "" {

		//执行获取流程
		deviceName, deviceNameMap, err := syncDevicesForSummary(c, homeId)
		if err != nil {
			c.Error(err)
			return nil, nil, err
		}

		fmt.Println("-----------deviceName-------", deviceName)
		fmt.Println("-----------deviceNameMap-------", deviceNameMap)

		err = db.GRedis.Set(
			context.Background(),
			redisKeyTuyaSummaryDeviceName+homeId,
			x.MustMarshal2String(deviceName),
			time.Hour*2).Err()
		if err != nil {
			c.Error(err)
			return nil, nil, err
		}

		err = db.GRedis.Set(
			context.Background(),
			redisKeyTuyaSummaryDeviceNameMap+homeId,
			x.MustMarshal2String(deviceNameMap),
			time.Hour*2).Err()
		if err != nil {
			c.Error(err)
			return nil, nil, err
		}

		return deviceName, deviceNameMap, nil
	}

	var devicesName []string
	err = x.MustUnmarshal([]byte(devicesNameResult), &devicesName)
	if err != nil {
		c.Error(err)
		return nil, nil, err
	}

	var devicesNameMap map[string]*device
	err = x.MustUnmarshal([]byte(deviceNameMapResult), &devicesNameMap)
	if err != nil {
		c.Error(err)
		return nil, nil, err
	}

	return devicesName, devicesNameMap, nil
}

// 分词，获取action信息
func getSummaryInfo(c *ava.Context, input string, devicesName []string) (*summaries, error) {

	mcList := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(fmt.Sprintf(summaryActionPrompts, x.MustMarshal2String(devicesName)))},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart(input)},
		},
	}

	//
	var resp summaries
	err := generateContentWithout(c, mcList, &resp)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	c.Debugf("getSummaryInfo |data=%v", x.MustMarshal2String(resp))
	return &resp, nil
}

var summaryActionPrompts = `根据我的意图描述，如果有多个动作意图，拆分出来，并找出你将要控制的设备，严格按照json数据格式返回给我，json数据前后不要出现任何字符；
### 设备列表：%s
### 返回json数据格式：
{
  "result": [
    {
      "content":"将客厅灯光调到4000k",
      "summary": "灯调光",
      "devices": [
         "客厅zigbee双色灯",
         "客厅双色1号温明装射灯"
      ]
    }
  ]
}

### 字段说明
content:完整的意图，例如：将客厅灯光调到4000k
summary: 简要意图，不超过5个字，例如：打开灯`

var summaryCommandPrompts = `根据我的意图描述，选择指令返回给我；
### 设备名称：%s
### 指令数据：%s
### 返回json数据格式：
{
	"success_msg":"[设备名称]已调到400k",
	"result":[{"code":"",value:400}]
}

### 字段说明
success_msg:设备控制结果，[设备名称]+结果，例如:当设备名称是“客厅一号灯”时，success_msg的值是: 客厅一号灯已打开`

var callbackAgentPrompts = `根据我的意图，从以下所有功能中找出最合适一个返回；
### 功能数据：%s
### 返回json数据格式:
{
	"argument":"summary"
}
`
