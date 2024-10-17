package langchain

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"vinesai/internel/ava"
	"vinesai/internel/db"
	"vinesai/internel/lib/tuyago"
	"vinesai/internel/x"

	"github.com/redis/go-redis/v9"
	"github.com/tmc/langchaingo/llms"
)

// 直接设备控制
type summary struct{ CallbacksHandler LogHandler }

func (s *summary) Name() string {
	return "summary"
}

func (s *summary) Description() string {
	return `意图描述是明确直接控制智能家居设备`
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

	var msg = "请告诉我你要控制什么设备"
	defer func() {
		setSummaryMsg(c, msg)
	}()

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

	if summary.FailureMsg != "" {
		return "", doneExitError
	}

	msg, err = chooseAndControlDevices(c, summary, devicesNameMap)
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

	return msg, doneExitError
	//return msg, nil
}

// 根据位置信息选取设备
func chooseAndControlDevices(c *ava.Context, s *summaries, devicesMap map[string]*device) (string, error) {

	var failureMessageArr []string
	var successMessageArr []string
	var offlineMessageArr []string
	var alreadyMessageArr []string

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

			if fs.FailureMsg != "" {
				failureMessageArr = append(failureMessageArr, fs.FailureMsg)
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
				offlineMessageArr = append(offlineMessageArr, offlineMsg(d.Name))
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
							c.Debugf("src=%s |dst=%s", x.MustMarshal2String(tmpFs), x.MustMarshal2String(tmpDeviceStatus))
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
				alreadyMessageArr = append(alreadyMessageArr, alreadyMsg(d.Name))
				continue
			}

			//执行设备控制
			var controlResp summaryControlDeviceResp
			//执行控制
			err = tuyago.Post(c, fmt.Sprintf("/v1.0/devices/%s/commands", d.Id), &summaryControlData{Commands: x.MustMarshal2String(fs.Result)}, &controlResp)

			if err != nil {
				c.Error(err)
				failureMessageArr = append(failureMessageArr, failureMsg(name))
				continue
			}

			if controlResp.Result && controlResp.Success {
				//判断语气中是否包含设备名称，这种情况是通过ai获取的结果
				if strings.Contains(fs.SuccessMsg, d.Name) {
					successMessageArr = append(successMessageArr, fs.SuccessMsg)
					fs.SuccessMsg = strings.Trim(fs.SuccessMsg, d.Name)
				} else {
					successMessageArr = append(successMessageArr, d.Name+fs.SuccessMsg)
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
				failureMessageArr = append(failureMessageArr, failureMsg(name))
			}
		}
	}

	var msg string

	if len(failureMessageArr) > 3 {
		msg = "有大量设备控制失败，请检查"
		return msg, nil
	}

	if len(successMessageArr) > 3 || len(alreadyMessageArr) > 3 {
		msg = "好的主人，设备都已控制成功啦"
		return msg, nil
	}

	if len(offlineMessageArr) > 3 {
		msg = "有大量设备离线，请检查"
		return msg, nil
	}

	for i := range failureMessageArr {
		msg += failureMessageArr[i] + ","
	}

	for i := range successMessageArr {
		msg += successMessageArr[i] + ","
	}

	for i := range offlineMessageArr {
		msg += offlineMessageArr[i] + ","
	}

	for i := range alreadyMessageArr {
		msg += alreadyMessageArr[i] + ","
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

	err = GenerateContentWithout(c, mcList, &resp)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	if len(resp.Result) == 0 {
		return nil, errors.New("no command response")
	}

	return &resp, nil
}

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
	err := GenerateContentWithout(c, mcList, &resp)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	c.Debugf("getummaryInfo |data=%v", x.MustMarshal2String(resp))
	return &resp, nil
}

var summaryActionPrompts = `根据我的意图描述，如果有多个动作意图，拆分出来，并找出你将要控制的设备，严格按照json数据格式返回给我，json数据前后不要出现任何字符；
### 设备列表：%s
### 返回json数据格式：
{
  "failure_msg":"请告诉我你想要控制什么设备",
  "result": [
    {
      "content":"将客厅灯光调到4000k",
      "summary": "调灯光",
      "devices": [
         "客厅zigbee双色灯",
         "客厅双色1号温明装射灯"
      ]
    }
  ]
}

### 字段说明
summary: 简要意图，不超过5个字，例如：打开灯，色温100，亮度4000,等等，如果有数值，则必须在该字段中包含;
content:完整的意图，例如：将客厅灯光调到4000k;
failure_msg:1.根据意图，如果分析意图失败，返回例子：请告诉我你想要控制什么设备;2.根据[设备列表]数据，如果没有找到设备，返回例子：你还没有空调

### 注意事项
1.如果设备没有明确关联，不要去控制其他设备，比如客厅有插排，我的意图是关灯，但是你不知道插排是不是控制灯的时候，就不要去关闭插排`

var summaryCommandPrompts = `根据我的意图描述，选择指令返回给我；
### 设备名称：%s
### 指令数据：%s
### 返回json数据格式：
{
	"success_msg":"[设备名称]已调到400",
	"failure_msg":"[设备名称]灯光色温最大值是1000",
	"result":[{"code":"",value:400}]
}

### 字段说明
success_msg:设备控制结果，[设备名称]+结果，例如:当设备名称是“客厅一号灯”时，success_msg的值是: 客厅一号灯已打开；
failure_msg:结合[指令数据]和我的意图分析是否存在指令或是否超出范围，指令正确时，这个字段为空`
