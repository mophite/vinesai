package langchain

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
	"vinesai/internel/ava"
	"vinesai/internel/db"
	"vinesai/internel/lib/tuyago"
	"vinesai/internel/x"

	"github.com/panjf2000/ants/v2"
	"github.com/redis/go-redis/v9"
	"github.com/tmc/langchaingo/jsonschema"
	"github.com/tmc/langchaingo/llms"
)

// 一次多个动作
type Actions struct {
	Result []*ActionData `json:"result"`
}

// eg.打开-客厅-灯
// key=action_productid_value_desc_specific_value
// 直接找出具体指令
type ActionData struct {
	Action       string `json:"action"`        //动作
	Position     string `json:"position"`      //位置，逗号分割，例如：“客厅，所有”，“所有”
	PositionType int    `json:"position_type"` //位置类型，0:无位置，1.房间位置，2.所有/周围/最近/全部/一些/部分/少许/百分之/一半，此种场景将种类下所有设备发送给ai
	DeviceName   string `json:"device_name"`   //设备名称
	Category     string `json:"category"`      //种类 dj代表灯具,逗号分割
	CategoryType int    `json:"category_type"` //1.总的设备,客厅灯，客厅吸顶灯，2.模糊名字的设备,一号灯
	ValueDesc    string `json:"value_desc"`    //值描述
	Value        string `json:"value"`         //具体值
	//ValueType     int    `json:"value_type"`     //1.数字值，2.非数字值
}

//type ActionCommand struct {
//	Result []*ActionCommandData `json:"Result"`
//}

// eg.客厅-灯-指令
type ActionCommandData struct {
	ProductId string      `json:"product_id"`
	Category  string      `json:"category"`
	Commands  []*function `json:"commands"`
}

// action+productId+valueDesc
type action struct {
	CallbacksHandler LogHandler
}

func (a *action) Name() string {
	return "action"
}

func (a *action) Description() string {
	return "通过用户意图，包含直接动作和目标，控制智能家居设备，例如：打开客厅灯"
}

func (a *action) Call(ctx context.Context, input string) (string, error) {
	var c = fromCtx(ctx)

	//获取用户所有位置，设备的category
	categoryList, err := getActionCategoryAndDevices(c)
	if err != nil {
		c.Error(err)
		return "", err
	}

	input = getFirstInput(c)
	//ai语义分词
	actions, err := getActionInfo(c, input, categoryList)
	if err != nil {
		c.Error(err)
		return "", err
	}

	msg, err := chooseAndControlActionDevices(c, actions, input)
	if err != nil {
		c.Error(err)
		return "", err
	}

	c.Debug("-------msg----", msg)

	return msg, nil
}

type ShortActionDeviceInfo struct {
	Result []*ShortActionDevice `json:"result"`
}

type ShortActionDevice struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Position  string `json:"position"`
	ProductId string `json:"product_id"`
	Category  string `json:"category"`
}

// 根据位置信息选取设备
func chooseAndControlActionDevices(c *ava.Context, actions *Actions, input string) (string, error) {

	var allDevices map[string][]*ShortActionDevice
	deviceStr, err := db.GRedis.Get(context.Background(), defaultActionShortDevicesKey+getHomeId(c)).Result()

	if err != nil {
		c.Error(err)
		return "", err
	}

	if deviceStr == "" {
		return "", errors.New("action short device not found")
	}

	err = x.MustNativeUnmarshal([]byte(deviceStr), &allDevices)
	if err != nil {
		c.Error(err)
		return "", err
	}

	var msg string

	for i := range actions.Result {
		data := actions.Result[i]
		//直接从房间获取设备
		if data.PositionType == 1 && data.CategoryType == 1 {

			var devices []*ShortActionDevice
			for k, v := range allDevices {
				if strings.Contains(k, data.Position) {
					devices = v
					break
				}
			}

			if len(devices) == 0 {
				continue
			}

			for ii := range devices {

				d := devices[ii]

				//过滤掉类型
				if !strings.Contains(data.Category, d.Category) {
					continue
				}

				fs, err := getActionCommand(c, data, d)
				if err != nil {
					c.Error(err)
					continue
				}

				var controlResp controlDeviceResp
				//执行控制
				err = tuyago.Post(c, fmt.Sprintf("/v1.0/devices/%s/commands", d.Id), &controlData{Commands: fs}, controlResp)

				if err != nil {
					c.Error(err)
					msg += failureMessage(data)
					continue
				}

				if controlResp.Result && controlResp.Success {
					msg += successMessage(data)
					//缓存指令
					err = db.GRedis.Set(context.Background(), getActionCategoryListKey(data.Category, d.ProductId, data.Action, data.ValueDesc, data.Value), x.MustMarshal2String(fs), 0).Err()
					if err != nil {
						c.Error(err)
					}
				} else {
					msg += failureMessage(data)
				}

			}

			continue
		}

		//todo 暂时用总类，可以继续细分，减少token数量
		if data.PositionType != 1 || data.CategoryType != 1 {

			var shortDevices ShortActionDeviceInfo

			//将数据发送给ai，让ai返回需要控制的设备
			mcList := []llms.MessageContent{
				{
					Role:  llms.ChatMessageTypeSystem,
					Parts: []llms.ContentPart{llms.TextPart(fmt.Sprintf(getActionDevicesCommand, x.MustMarshal2String(allDevices)))},
				},
				{
					Role:  llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{llms.TextPart(input)},
				},
			}

			//
			result, err := generateContent(mcList, shortActionDeviceTool)
			if err != nil {
				c.Error(err)
				return "", err
			}

			if len(result.Choices) == 0 || result.Choices[0].FuncCall == nil || result.Choices[0].FuncCall.Arguments == "" {
				return "", fmt.Errorf("ai no resp |result=%v", result)
			}

			err = x.MustNativeUnmarshal([]byte(result.Choices[0].FuncCall.Arguments), &shortDevices)
			if err != nil {
				c.Error(err)
				return "", err
			}

			if len(shortDevices.Result) == 0 {
				return "", errors.New("no device need be control")
			}

			for i := range shortDevices.Result {
				d := shortDevices.Result[i]

				//获取设备指令
				fs, err := getActionCommand(c, data, d)
				if err != nil {
					c.Error(err)
					continue
				}

				var controlResp controlDeviceResp
				//执行控制
				err = tuyago.Post(c, fmt.Sprintf("/v1.0/devices/%s/commands", d.Id), &controlData{Commands: fs}, controlResp)

				if err != nil {
					c.Error(err)
					msg += failureMessage(data)
					continue
				}

				if controlResp.Result && controlResp.Success {
					msg += successMessage(data)
					//缓存指令
					err = db.GRedis.Set(context.Background(), getActionCategoryListKey(data.Category, d.ProductId, data.Action, data.ValueDesc, data.Value), x.MustMarshal2String(fs), 0).Err()
					if err != nil {
						c.Error(err)
					}
				} else {
					msg += failureMessage(data)
				}
			}

			continue
		}

		////todo 具体设备,例如：打开名称为“一号灯”的设备,或者打开名称为“客厅1号灯”
		//if data.CategoryType == 2 {
		//	//通过ai搜索到最相近的设备
		//
		//	continue
		//}
		//
		////todo 后期考虑是否还可以继续拆封
		////无位置或者特殊位置,例如：打开周围的灯或打开客厅“1号灯”
		//if data.PositionType == 2 || data.PositionType == 0 {
		//	//将所有设备发送给ai进行筛选
		//
		//	continue
		//}
	}

	return msg, nil
}

// 设备，动作，值
func successMessage(a *ActionData) string {
	if !strings.Contains(a.ValueDesc, "开关") {
		return a.Position + a.DeviceName + "已" + a.Action + "为" + a.ValueDesc + a.Value
	}

	return a.Position + a.DeviceName + "已" + a.Action
}

func failureMessage(a *ActionData) string {
	if !strings.Contains(a.ValueDesc, "开关") {
		return a.Position + a.DeviceName + a.Action + "为" + a.ValueDesc + a.Value + "失败"
	}

	return a.Position + a.DeviceName + a.Action + "失败"
}

type controlDeviceResp struct {
	Result  bool   `json:"result"`
	Success bool   `json:"success"`
	T       int    `json:"t"`
	Tid     string `json:"tid"`
}

type controlData struct {
	Commands []*function `json:"commands"`
}

func getActionCommand(c *ava.Context, action *ActionData, s *ShortActionDevice) ([]*function, error) {
	var fs []*function
	var key = getActionCategoryListKey(s.Category, s.ProductId, action.Action, action.ValueDesc, action.Value)
	//先从redis获取设备指令
	result, err := db.GRedis.Get(context.Background(), key).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		c.Error(err)
		return nil, err
	}

	//有些设备没有指令
	if result == "-" {
		return nil, fmt.Errorf("this productID=%s category=%s no function", s.ProductId, s.Category)
	}

	if result == "" {
		//从涂鸦获取设备指令
		var cmdResp = &commandsResp{}
		//从涂鸦api查询指令
		err = tuyago.Get(c, fmt.Sprintf("/v1.0/devices/functions?device_ids=%s", s.Id), cmdResp)

		if err != nil {
			c.Errorf("productId=%s |deviceId=%s |err=%v", s.ProductId, s.Category, err)
			return nil, err
		}

		if cmdResp.Success && len(cmdResp.Result) == 0 {
			err = db.GRedis.Set(context.Background(), key, "-", 0).Err()
			if err != nil {
				ava.Error(err)
				return nil, err
			}

			c.Errorf("this productID=%s category=%s no function", s.ProductId, s.Category)
			return nil, fmt.Errorf("this productID=%s category=%s no function", s.ProductId, s.Category)
		}

		//发送给ai请求获取指令
		var sendData = &ActionCommandData{
			ProductId: s.ProductId,
			Category:  s.Category,
			Commands:  cmdResp.Result[0].Functions,
		}
		//将数据发送给ai，让ai返回控制指令
		mcList := []llms.MessageContent{
			{
				Role:  llms.ChatMessageTypeSystem,
				Parts: []llms.ContentPart{llms.TextPart(fmt.Sprintf(getActionProductIdCommand, x.MustMarshal2String(&sendData)))},
			},
			{
				Role:  llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{llms.TextPart(x.MustMarshal2String(action))},
			},
		}

		//
		result, err := langchaingoOpenAi.GenerateContent(
			context.Background(),
			mcList,
			llms.WithTemperature(0.5),
			llms.WithN(1),
			llms.WithTopP(0.5),
			llms.WithTools(actionCommandDataTool),
		)
		if err != nil {
			c.Error(err)
			return nil, err
		}

		if len(result.Choices) == 0 || result.Choices[0].FuncCall == nil || result.Choices[0].FuncCall.Arguments == "" {
			return nil, fmt.Errorf("ai no resp |result=%v", result)
		}

		var resp ActionCommandData
		err = x.MustNativeUnmarshal([]byte(result.Choices[0].FuncCall.Arguments), &resp)
		if err != nil {
			c.Error(err)
			return nil, err
		}

		fs = resp.Commands

	}

	return fs, nil
}

type NameDevices struct {
	Result []*NameDevice `json:"result"`
}

type NameDevice struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

// 同步用户设备数据
// 返回所有设备，以及设备种类和种类描述
func syncDevicesFromAction(c *ava.Context, r []*room) (map[string][]*ShortActionDevice, map[string]*ActionCategories, map[string][]*NameDevice, error) {

	//返回的设备信息，key：位置position
	var PositionRoomShortDevicesMap = make(map[string][]*ShortActionDevice, 2)
	//用于指定设备查询,key:位置position
	var nameShortDevices = make(map[string][]*NameDevice, 2)
	//用户所有设备类型，key:位置position
	var categoriesList = make(map[string]*ActionCategories, 2)

	var mux = new(sync.Mutex)

	pool, err := ants.NewPool(len(r))
	if err != nil {
		c.Error(err)
		return nil, nil, nil, err
	}

	//遍历房间获取设备
	for i := range r {
		var tmpRoom = r[i]
		err = pool.Submit(func() {
			var tmpResp = &deviceListResp{}

			err := tuyago.Get(c, fmt.Sprintf("/v1.0/homes/%s/rooms/%d/devices", getHomeId(c), tmpRoom.RoomID), tmpResp)

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

			c.Debugf("所有设备 ｜tmpResp=%v", x.MustMarshal2String(tmpResp))

			//用户判断category是否重复
			var skipCategoryMap = make(map[string]bool)
			var categoryData []*actionCategoryData
			var position = tmpRoom.Name + strconv.FormatInt(tmpRoom.RoomID, 10)
			var tmpShortDevice []*ShortActionDevice
			var tmpNameShortDevice []*NameDevice

			mux.Lock()
			defer mux.Unlock()

			for ii := range tmpResp.Result {
				tmpData := tmpResp.Result[ii]

				categoryName := getCategoryName(tmpData.Category)

				if categoryName == "" {
					c.Debugf("category name not exist |category=%v", tmpData.Category)
					continue
				}

				var short = &ShortActionDevice{
					Id:        tmpData.Id,
					Name:      tmpData.Name,
					Category:  tmpData.Category,
					ProductId: tmpData.ProductId,
					Position:  position,
				}
				tmpShortDevice = append(tmpShortDevice, short)
				tmpNameShortDevice = append(tmpNameShortDevice, &NameDevice{
					Id:   tmpData.Id,
					Name: tmpData.Name,
				})

				if _, ok := skipCategoryMap[categoryName]; !ok {
					skipCategoryMap[categoryName] = true

					categoryData = append(categoryData, &actionCategoryData{
						CategoryName: categoryName,
						Category:     tmpData.Category,
					})
				}
			}

			if len(tmpShortDevice) == 0 || len(tmpNameShortDevice) == 0 {
				return
			}

			PositionRoomShortDevicesMap[position] = tmpShortDevice
			nameShortDevices[position] = tmpNameShortDevice
			categoriesList[position] = &ActionCategories{CategoryData: categoryData}
		})
	}

	err = pool.ReleaseTimeout(time.Second * 10)
	if err != nil {
		c.Error(err)
		return nil, nil, nil, err
	}

	return PositionRoomShortDevicesMap, categoriesList, nameShortDevices, nil
}

// 单个房间里的设备类型
type ActionCategories struct {
	CategoryData []*actionCategoryData `json:"category_data"`
}

type actionCategoryData struct {
	CategoryName string `json:"category_name"`
	Category     string `json:"category"`
}

//key:roomName+roomId
func getActionCategoryAndDevices(c *ava.Context) (map[string]*ActionCategories, error) {
	result, err := db.GRedis.Get(context.Background(), defaultActionCategoryListKey+getHomeId(c)).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		c.Error(err)
		return nil, err
	}

	if result == "" || errors.Is(err, redis.Nil) {
		//获取房间信息
		var roomResp = &roomInfo{}

		err = tuyago.Get(c, fmt.Sprintf("/v1.0/homes/%s/rooms", getHomeId(c)), roomResp)

		if err != nil {
			c.Error(err)
			return nil, err
		}

		if !roomResp.Success {
			return nil, errors.New("获取房间信息失败")
		}

		c.Debugf("roomInfo: %v", roomResp)

		if len(roomResp.Result.Rooms) == 0 {
			return nil, errors.New("该用户没有任何房间")
		}

		//执行获取流程
		shortDeviceList, categoryList, nameShortDevice, err := syncDevicesFromAction(c, roomResp.Result.Rooms)
		if err != nil {
			c.Error(err)
			return nil, err
		}

		fmt.Println("-----------short-------", len(x.MustMarshal2String(shortDeviceList)))

		err = db.GRedis.Set(
			context.Background(),
			defaultActionCategoryListKey+getHomeId(c),
			x.MustMarshal2String(categoryList),
			time.Hour*2).Err()
		if err != nil {
			c.Error(err)
			return nil, err
		}

		err = db.GRedis.Set(
			context.Background(),
			defaultActionShortDevicesKey+getHomeId(c),
			x.MustMarshal2String(shortDeviceList),
			time.Hour*2).Err()
		if err != nil {
			c.Error(err)
			return nil, err
		}

		err = db.GRedis.Set(
			context.Background(),
			defalutActionNameDeviceKey+getHomeId(c),
			x.MustMarshal2String(nameShortDevice),
			time.Hour*2).Err()
		if err != nil {
			c.Error(err)
			return nil, err
		}

		return categoryList, nil
	}

	var categoryList map[string]*ActionCategories
	err = x.MustUnmarshal([]byte(result), &categoryList)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	return categoryList, nil
}

// 分词，获取action信息
func getActionInfo(c *ava.Context, input string, categoryList map[string]*ActionCategories) (*Actions, error) {

	mcList := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(fmt.Sprintf(segmentationInput, x.MustMarshal2String(categoryList)))},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart(input)},
		},
	}

	//
	result, err := langchaingoOpenAi.GenerateContent(
		context.Background(),
		mcList,
		llms.WithTemperature(0.5),
		llms.WithN(1),
		llms.WithTopP(0.5),
		llms.WithTools(actionsTool),
	)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	if len(result.Choices) == 0 || result.Choices[0].FuncCall == nil || result.Choices[0].FuncCall.Arguments == "" {
		return nil, fmt.Errorf("ai no resp |result=%v", result)
	}

	c.Debugf("ai分词获取动作 |result=%v", result.Choices[0].FuncCall.Arguments)

	var resp Actions

	err = x.MustNativeUnmarshal([]byte(result.Choices[0].FuncCall.Arguments), &resp)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	if len(resp.Result) == 0 {
		return nil, fmt.Errorf("ai no resp |resp=%v", resp)
	}

	c.Debugf("actions: %v", x.MustMarshal2String(&resp))

	return &resp, nil
}

var actionsTool = []llms.Tool{
	{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "execute_actions",
			Description: "执行多个动作，根据提供的参数控制设备。",
			Parameters: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"result": {
						Type:        jsonschema.Array,
						Description: "多个动作的详细信息",
						Items: &jsonschema.Definition{
							Type: jsonschema.Object,
							Properties: map[string]jsonschema.Definition{
								"action": {
									Type:        jsonschema.String,
									Description: "动作，例如打开、关闭、调亮等",
								},
								"position": {
									Type:        jsonschema.String,
									Description: "位置，例如客厅，所有",
								},
								"position_type": {
									Type:        jsonschema.Integer,
									Description: "位置类型，0.无任何位置，1.房间位置，2.所有/周围/最近/全部等",
								},
								"device_name": {
									Type:        jsonschema.String,
									Description: "从我的意图中得到设备名称，“灯",
								},
								"category": {
									Type:        jsonschema.String,
									Description: "设备种类，逗号分割；例如'dj,xdd'",
								},
								"category_type": {
									Type:        jsonschema.Integer,
									Description: "设备种类：1.所有/全部/某一类型，2.具体到某个设备，例如：走廊第一个吊灯",
								},
								"value_desc": {
									Type:        jsonschema.String,
									Description: "值描述，例如：倒计时",
								},
								"value": {
									Type:        jsonschema.String,
									Description: "具体值，例如：3600",
								},
							},
							Required: []string{"action", "position", "position_type", "device_name", "category", "category_type", "value_desc", "value"},
						},
					},
				},
				Required: []string{"result"},
			},
		},
	},
}

var shortActionDeviceTool = []llms.Tool{
	{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "get_short_action_devices",
			Description: "根据我的意图获取需要控制的设备数据信息。",
			Parameters: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"result": {
						Type:        jsonschema.Array,
						Description: "设备动作信息列表",
						Items: &jsonschema.Definition{
							Type: jsonschema.Object,
							Properties: map[string]jsonschema.Definition{
								"id": {
									Type:        jsonschema.String,
									Description: "设备ID",
								},
								"category": {
									Type:        jsonschema.String,
									Description: "产品类型",
								},
								"product_id": {
									Type:        jsonschema.String,
									Description: "产品ID",
								},
							},
							Required: []string{"id", "category", "product_id"},
						},
					},
				},
				Required: []string{"result"},
			},
		},
	},
}

var actionCommandDataTool = []llms.Tool{
	{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "get_action_command_data",
			Description: "获取设备的动作命令数据，包括产品ID、类别和可用命令。",
			Parameters: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"product_id": {
						Type:        jsonschema.String,
						Description: "产品ID",
					},
					"category": {
						Type:        jsonschema.String,
						Description: "设备类别",
					},
					"commands": {
						Type:        jsonschema.Array,
						Description: "可用命令列表",
						Items: &jsonschema.Definition{
							Type: jsonschema.Object,
							Properties: map[string]jsonschema.Definition{
								"code": {
									Type:        jsonschema.String,
									Description: "命令代码",
								},
								"values": {
									Type:        jsonschema.Object,
									Description: "命令值（可以是任何类型）",
								},
							},
							Required: []string{"code", "values"},
						},
					},
				},
				Required: []string{"product_id", "category", "commands"},
			},
		},
	},
}
