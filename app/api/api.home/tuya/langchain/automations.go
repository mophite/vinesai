package langchain

import (
	"context"
	"fmt"
	"time"
	"vinesai/internel/ava"
	"vinesai/internel/db"
	"vinesai/internel/langchaingo/llms"
	"vinesai/internel/lib/tuyago"
	"vinesai/internel/x"

	"github.com/panjf2000/ants/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// 自动化创建
// 安防，传感器触发，定时，延时，天气，温度等等复杂条件
type automations struct{ CallbacksHandler LogHandler }

func (a *automations) Name() string {
	return "smart_home_automations"
}

func (a *automations) Description() string {
	return "智能家居自动化创建，涉及条件：设备状态改变、天气指标、定时执行、固定时间执行的自动化"
}

// 条件匹配类型：任意条件满足触发，全部条件满足触发
func (a *automations) Call(ctx context.Context, input string) (string, error) {
	var c = fromCtx(ctx)
	var homeId = getHomeId(c)
	input = getFirstInput(c)

	//获取用户位置信息
	var h homeInfo
	var msg = "请更详细说明你想创建的自动化"
	err := db.Mgo.Collection(mgoCollectionUser).FindOne(context.Background(), bson.M{"_id": homeId}).Decode(&h)
	if err != nil {
		c.Error(err)
		return msg, err
	}

	pool, err := ants.NewPool(10)
	if err != nil {
		c.Error(err)
		return "ants no submit", err
	}

	var actionResp string
	var weatherResp []*condition
	var deviceStatusResp []*condition
	var timingResp []*timing
	var preResp []*preconditions

	_ = pool.Submit(func() {
		actionResp, err = a.actions(c, homeId, input)
		if err != nil {
			c.Error(err)
			return
		}
	})

	_ = pool.Submit(func() {
		weatherResp, err = a.WeatherCondition(c, h.Lat, h.Lon, input)
		if err != nil {
			c.Error(err)
			return
		}
	})

	_ = pool.Submit(func() {
		deviceStatusResp, err = a.deviceStatusCondition(c, homeId, input)
		if err != nil {
			c.Error(err)
			return
		}
	})

	_ = pool.Submit(func() {
		timingResp, err = a.timingCondition(c, input)
		if err != nil {
			c.Error(err)
			return
		}
	})
	_ = pool.Submit(func() {
		preResp, err = a.fixedTimeCondition(c, input)
		if err != nil {
			c.Error(err)
			return
		}
	})

	err = pool.ReleaseTimeout(time.Minute * 2)
	if err != nil {
		c.Error(err)
		return msg, err
	}

	//通过ai返回创建一键场景的数据
	mcList1 := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(fmt.Sprintf(
				createAutoPrompts,
				x.MustMarshal2String(weatherResp),
				x.MustMarshal2String(deviceStatusResp),
				x.MustMarshal2String(timingResp),
				x.MustMarshal2String(preResp),
				x.MustMarshal2String(actionResp),
			))},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart(input)},
		},
	}

	var createAuto createAutoData

	err = GenerateContentWithout(c, mcList1, &createAuto)
	if err != nil {
		c.Error(err)
		return msg, err
	}

	createAuto.Background = defaultBackgroudPicture

	//发起自动化创建
	var createAutoResp struct {
		Success bool   `json:"success"`
		T       int64  `json:"t"`
		Result  string `json:"result"` //返回场景id
		Tid     string `json:"tid"`
	}

	err = tuyago.Post(c, fmt.Sprintf("/v1.0/homes/%s/automations", homeId), &createAuto, &createAutoResp)
	if err != nil {
		c.Error(err)
		return msg, err
	}

	return "创建" + createAuto.Name + "成功了", nil
}

type condition struct {
	Display struct {
		Code     string `json:"code"`
		Operator string `json:"operator"`
		Value    string `json:"value"`
	} `json:"display"`
	EntityID   string `json:"entity_id"`
	EntityType int    `json:"entity_type"`
	OrderNum   int    `json:"order_num"`
}

// 天气条件
func (a *automations) WeatherCondition(c *ava.Context, lat, lon, input string) ([]*condition, error) {
	//根据设备经纬度查询城市id
	var cityResp struct {
		CityId string `json:"city_id"`
	}
	err := tuyago.Get(c, fmt.Sprintf("/v1.0/iot-03/cities/positions?lon=%s&lat=%s", lon, lat), &cityResp)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	mcList1 := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(fmt.Sprintf(autoWeatherConditionPrompts, WeatherCode))},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart(input)},
		},
	}

	var cond []*condition

	err = GenerateContentWithout(c, mcList1, &cond)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	for i := range cond {
		cond[i].EntityType = 3
		cond[i].EntityID = cityResp.CityId
	}

	return cond, nil
}

// 设备状态改变
func (a *automations) deviceStatusCondition(c *ava.Context, homeId, input string) ([]*condition, error) {
	//查询出所有status有值的联动条件数据
	var statusCodeResp []*homeStatus
	cur, err := db.Mgo.Collection(mgoCollectionCodes).Find(context.Background(), bson.M{"homeid": homeId})
	if err != nil {
		c.Error(err)
		return nil, err
	}

	if err := cur.All(context.Background(), &statusCodeResp); err != nil {
		c.Error(err)
		return nil, err
	}

	var statusCodeResult = make([]*homeStatus, 0, len(statusCodeResp))
	for i := range statusCodeResult {
		if len(statusCodeResult[i].Status) == 0 {
			continue
		}
		statusCodeResult = append(statusCodeResult, statusCodeResult[i])
	}

	//获取设备状态条件
	mcList1 := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(fmt.Sprintf(autoStatusDeviceConditionPrompts, x.MustMarshal2String(&statusCodeResult)))},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart(input)},
		},
	}

	var cond []*condition

	err = GenerateContentWithout(c, mcList1, &cond)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	for i := range cond {
		cond[i].EntityType = 3
	}

	return cond, nil
}

// 定时条件
type timing struct {
	Display struct {
		Date       string `json:"date"`
		Loops      string `json:"loops"`
		Time       string `json:"time"`
		TimezoneID string `json:"timezone_id"`
	} `json:"display"`
	EntityID   string `json:"entity_id"`
	EntityType int    `json:"entity_type"`
	OrderNum   int    `json:"order_num"`
}

func (a *automations) timingCondition(c *ava.Context, input string) ([]*timing, error) {
	//获取设备状态条件
	mcList1 := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(autoTimingConditionPrompts)},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart(input)},
		},
	}

	var cond []*timing

	err := GenerateContentWithout(c, mcList1, &cond)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	for i := range cond {
		cond[i].EntityType = 6
		cond[i].EntityID = "timer"
		cond[i].Display.TimezoneID = "Asia/Shanghai"
	}

	return cond, nil
}

// 固定时间条件
type preconditions struct {
	CondType string `json:"cond_type"`
	Display  struct {
		//CityID        string `json:"cityId"`
		//CityName      string `json:"cityName"`
		End   string `json:"end"`
		Loops string `json:"loops"`
		Start string `json:"start"`
		//TimeInterval  string `json:"timeInterval"`
		//TimeZoneID string `json:"timeZoneId"`
		//TimeInterval0 string `json:"time_interval"`
		TimezoneID string `json:"timezone_id"`
	} `json:"display"`
	//ID string `json:"id"`
}

func (a *automations) fixedTimeCondition(c *ava.Context, input string) ([]*preconditions, error) {
	//获取设备状态条件
	mcList1 := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(autoFixTimeConditionPrompts)},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart(input)},
		},
	}

	var cond []*preconditions

	err := GenerateContentWithout(c, mcList1, &cond)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	for i := range cond {
		cond[i].CondType = "timeCheck"
		cond[i].Display.TimezoneID = "Asia/Shanghai"
	}

	return cond, nil
}

// 动作指令执行
func (a *automations) actions(c *ava.Context, homeId, input string) (string, error) {

	var msg = "请告诉我更详细的规则"

	//获取支持场景的设备列表
	var ssResp supportSceneDevices
	err := tuyago.Get(c, fmt.Sprintf("/v1.0/homes/%s/scene/devices", homeId), &ssResp)
	if err != nil {
		c.Error(err)
		return msg, err
	}

	if len(ssResp.Result) == 0 {
		return "你没有可以创建场景的设备", err
	}

	filter := bson.M{"homeid": homeId}

	//获取筛选后的设备支持的联动规则，指令
	cur, err := db.Mgo.Collection(mgoCollectionCodes).Find(context.Background(), filter)
	if err != nil {
		c.Error(err)
		return "开了点小差，重试一下", err
	}

	var codesResp []homeFunctionAndStatus
	err = cur.All(context.Background(), &codesResp)
	if err != nil {
		c.Error(err)
		return "开了点小差，重试一下", err
	}

	var codesReq = make([]homeFunction, 0, len(codesResp))

	for i := range codesResp {
		var code = codesResp[i]
		if len(code.Functions) == 0 {
			continue
		}
		codesReq = append(codesReq, homeFunction{
			DeviceId:  code.DeviceId,
			Functions: code.Functions,
			Name:      code.Name,
		})
	}

	//通过ai返回创建一键场景的数据
	mcList1 := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(fmt.Sprintf(autoActionCreatePrompts, removeWhitespace(x.MustMarshal2String(codesReq))))},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart(input)},
		},
	}

	var resultAction createSceneFromAiResp

	err = GenerateContentWithout(c, mcList1, &resultAction)
	if err != nil {
		c.Error(err)
		return msg, err
	}

	var actionsData createAuto
	actionsData.Actions = resultAction.Actions

	return x.MustMarshal2String(actionsData), nil
}

//位置tool：待定
