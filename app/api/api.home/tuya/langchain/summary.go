package langchain

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"
	"vinesai/internel/ava"
	"vinesai/internel/db"
	"vinesai/internel/lib/tuyago"
	"vinesai/internel/x"

	"vinesai/internel/langchaingo/llms"

	"github.com/panjf2000/ants/v2"
	"github.com/redis/go-redis/v9"
)

// 直接设备控制
type summary struct{ CallbacksHandler LogHandler }

func (s *summary) Name() string {
	return "summary"
}

func (s *summary) Description() string {
	return `明确直接控制智能家居设备`
}

func (s *summary) Call(ctx context.Context, input string) (string, error) {

	var c = fromCtx(ctx)
	var homeId = getHomeId(c)

	var msg = "请告诉我你要控制什么设备"

	//获取所有设备
	devicesName, devicesNameMap, err := getSummaryDevices(c, homeId)
	if err != nil {
		c.Error(err)
		return msg, err
	}

	summary, err := getSummaryInfo(c, getFirstInput(c), devicesName)
	if err != nil {
		c.Error(err)
		return msg, err
	}

	if summary.FailureMsg != "" {
		return summary.FailureMsg, nil
	}

	msg, err = chooseAndControlDevices(c, summary, devicesNameMap)
	if err != nil {
		c.Error(err)
		return msg, err
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

	//return msg, doneExitError
	return msg, nil
}

// 根据位置信息选取设备
func chooseAndControlDevices(c *ava.Context, s *summaries, devicesMap map[string]*mgoDocDevice) (string, error) {

	if len(devicesMap) == 0 {
		return "没有设备需要控制", errors.New("no mgoDocDevice need to control")
	}

	var failureMessageArr []string
	var successMessageArr []string
	var offlineMessageArr []string
	var alreadyMessageArr []string

	pool, err := ants.NewPool(5)
	if err != nil {
		c.Error(err)
		return "ants no submit", err
	}
	var mux = new(sync.Mutex)

	for i := range s.Result {
		summa := s.Result[i]
		ds := strings.Split(summa.Devices, ",")
		for ii := range ds {
			name := ds[ii]

			err = pool.Submit(func() {

				d, ok := devicesMap[name]
				if !ok {
					return
				}

				//从redis获取简短的设备指令
				fs, err := getSummaryCommand(c, d.Category, d.ProductID, summa.Summary, d.ID, d.Name, summa.Content)
				if err != nil {
					ava.Error(err)
					return
				}

				if fs.FailureMsg != "" {
					syncStringArr(mux, &failureMessageArr, fs.FailureMsg)
				}

				var tmpDevicesResp = &deviceResp{}

				err = tuyago.Get(c, fmt.Sprintf("/v1.0/devices/%s", d.ID), tmpDevicesResp)

				if err != nil {
					ava.Error(err)
					return
				}

				if !tmpDevicesResp.Success {
					c.Debugf("get mgoDocDevice list from room fail |data=%v |id=%v", tmpDevicesResp, d.ID)
					return
				}

				//判断设备状态是否在线
				if !tmpDevicesResp.Result.Online {
					syncStringArr(mux, &offlineMessageArr, offlineMsg(d.Name))
					return
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
					syncStringArr(mux, &alreadyMessageArr, alreadyMsg(d.Name))
					return
				}

				//执行设备控制
				var controlResp summaryControlDeviceResp
				//执行控制
				err = tuyago.Post(c, fmt.Sprintf("/v1.0/devices/%s/commands", d.ID), &summaryControlData{Commands: x.MustMarshal2String(fs.Result)}, &controlResp)

				if err != nil {
					ava.Error(err)
					syncStringArr(mux, &failureMessageArr, failureMsg(d.Name))
					return
				}

				if controlResp.Result && controlResp.Success {
					//判断语气中是否包含设备名称，这种情况是通过ai获取的结果
					if strings.Contains(fs.SuccessMsg, d.Name) {
						syncStringArr(mux, &successMessageArr, fs.SuccessMsg)

						fs.SuccessMsg = strings.Trim(fs.SuccessMsg, d.Name)
					} else {
						syncStringArr(mux, &successMessageArr, d.Name+fs.SuccessMsg)
					}

					//缓存指令
					err = db.GRedis.HSet(
						context.Background(),
						getSummaryCategoryListKey(d.Category, d.ProductID),
						summa.Summary,
						x.MustMarshal2String(fs)).Err()
					if err != nil {
						ava.Error(err)
					}
				} else {
					syncStringArr(mux, &failureMessageArr, failureMsg(name))
				}
			})

			if err != nil {
				ava.Error(err)
				continue
			}

		}
	}

	err = pool.ReleaseTimeout(time.Second * 20)
	if err != nil {
		c.Error(err)
		return "ants release timeout", err
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
		if i == len(failureMessageArr)-1 {
			msg += "控制失败"
		}
	}

	for i := range successMessageArr {
		msg += successMessageArr[i] + ","
	}

	for i := range offlineMessageArr {
		msg += offlineMessageArr[i] + ","
		if i == len(offlineMessageArr)-1 {
			msg += "已离线"
		}
	}

	for i := range alreadyMessageArr {
		msg += alreadyMessageArr[i] + ","
		if i == len(alreadyMessageArr)-1 {
			msg += "已经是你想要的状态了"
		}
	}

	if msg == "" {
		return "没有设备需要控制", nil
	}

	return msg, nil
}

func syncStringArr(mux *sync.Mutex, arr *[]string, message string) {
	mux.Lock()
	*arr = append(*arr, message)
	mux.Unlock()
}

// 设备，动作，值
func successMsg(name, successMsg string) string {
	return name + successMsg
}

func failureMsg(name string) string {
	return name
}

func offlineMsg(name string) string {
	return name
}

func alreadyMsg(name string) string {
	return name
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

func getSummaryDevices(c *ava.Context, homeId string) ([]string, map[string]*mgoDocDevice, error) {
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

	var devicesNameMap map[string]*mgoDocDevice
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

var summaryActionPrompts = `分析我的意图，找出你将要控制的设备，严格按照返回例子的JSON数据结构返回；
### 1.设备列表：%s

### 2.字段说明：
failure_msg:1.例子：你的要求暂时无法实现;2.根据[设备列表]数据，没有找到设备例子：你还没有空调
result: 对象数组，一句话里面可能有一个或多个意图
content:完整的意图，例如：将客厅灯光调到4000k;
summary:简要意图，不超过5个字，例如：打开灯，色温100，亮度4000等，如果有数值，则必须在该字段中包含;
devices:字符串，逗号分隔设备，如果可能有多个设备需要被控制，这些设备都写入到devices中

### 3.注意事项
如果开关、插座没有明确关联灯具，不要去使用开关、插座，除非我告诉你开关、插座是控制什么设备的

### 4.返回例子：
{
  "failure_msg":"请告诉我你想要控制什么设备",
  "result": [
    {
      "content":"将客厅灯光调到4000k",
      "summary":"调灯光",
      "devices":"客厅zigbee双色灯,客厅双色1号温明装射灯"
    }
  ]
}`

var summaryCommandPrompts = `根据我的意图描述，选择指令按照json格式返回给我；
### 设备名称：%s
### 指令数据：%s
### 返回例子：
{
	"success_msg":"[设备名称]已调到400",
	"failure_msg":"[设备名称]灯光色温最大值是1000",
	"result":[{"code":"",value:400}]
}

### 字段说明
success_msg:设备控制结果，[设备名称]+结果，例如:当设备名称是“客厅一号灯”时，success_msg的值是: 客厅一号灯已打开；
failure_msg:结合[指令数据]和我的意图分析是否存在指令或是否超出范围，指令正确时，这个字段为空`
