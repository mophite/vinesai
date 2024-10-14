package langchain

//获取设备信息，指令信息
//获取ai返回的设备控制指令

import (
	"context"
	"errors"
	"fmt"
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

type devicesAgent struct {
	CallbacksHandler LogHandler
}

func (d *devicesAgent) Name() string { return "devices_control" }

func (d *devicesAgent) Description() string {
	return "对智能家居设备控制很有用"
}

// 减少token的方法:
// 1.发送房间和设备种类给ai进行筛选
// 排除被动设备，无指令集设备
// 根据房间，设备种类来选择设备列表
// 按照productId将设备和指令集放在一起,减少指令集太多造成token过多
// 发给ai
func (d *devicesAgent) Call(ctx context.Context, input string) (string, error) {
	var c = fromCtx(ctx)

	input = getFirstInput(c)
	categoryList, err := getCategoryList(c)
	if err != nil {
		c.Error(err)
		return "", err
	}

	category, err := getNeedToControl(c, categoryList, input)
	if err != nil {
		c.Error(err)
		return "", err
	}

	result, err := controlDevices(c, category, input)
	if err != nil {
		c.Error(err)
		return "", err
	}

	c.Debugf("result=%v", result)

	return "", nil
}

//func (d *devicesAgent) Call(ctx context.Context, input string) (string, error) {
//	fmt.Println("------23---", input)
//	var c = fromCtx(ctx)
//
//	r, err := skipRoom(c, input)
//	if err != nil {
//		c.Error(err)
//		return "", err
//	}
//
//	var mux = new(sync.Mutex)
//	var deviceShortList []*shortDeviceListResp
//	var deviceList []*deviceListResp
//	//var productIdMap = make(map[string]string) //对应设备id,当查从redis查询不到指令时，用设备id从涂鸦获取
//	//var productId []string
//	var deviceFullInfoMap = make(map[string]*device, 10)
//
//	pool, err := ants.NewPool(len(r))
//	if err != nil {
//		c.Error(err)
//		return "", err
//	}
//
//	//遍历房间获取设备
//	for i := range r {
//		var tmpRoom = r[i]
//		err = pool.Submit(func() {
//			var tmpResp = &deviceListResp{}
//
//			err := tuyago.Get(c, fmt.Sprintf("/v1.0/homes/%s/rooms/%d/devices", homeId, tmpRoom.RoomID), tmpResp)
//
//			if err != nil {
//				ava.Error(err)
//				return
//			}
//
//			if !tmpResp.Success {
//				ava.Debugf("get device list from room fail |data=%v |id=%v", tmpResp, tmpRoom.RoomID)
//				return
//			}
//
//			if len(tmpResp.Result) == 0 {
//				return
//			}
//
//			//去掉离线设备
//			var tmpShortDevice []*shortDevice
//			var tmpDevice []*device
//
//			for ii := range tmpResp.Result {
//				tmpData := tmpResp.Result[ii]
//
//				if tmpData.Online {
//
//					tmpShortDevice = append(tmpShortDevice, &shortDevice{
//						Id:   tmpData.Id,
//						Name: tmpData.Name,
//					})
//					tmpDevice = append(tmpDevice, tmpData)
//
//					mux.Lock()
//					if _, ok := productIdMap[tmpData.ProductId]; !ok {
//						productId = append(productId, tmpData.ProductId)
//						productIdMap[tmpData.ProductId] = tmpData.Id
//					}
//					deviceFullInfoMap[tmpData.Id] = tmpData
//					mux.Unlock()
//				}
//			}
//
//			if len(tmpShortDevice) == 0 {
//				return
//			}
//
//			mux.Lock()
//			deviceList = append(deviceList, &deviceListResp{Result: tmpDevice, Position: tmpRoom.Name})
//			deviceShortList = append(deviceShortList, &shortDeviceListResp{Devices: tmpShortDevice, Position: tmpRoom.Name})
//			mux.Unlock()
//		})
//	}
//
//	err = pool.ReleaseTimeout(time.Second * 10)
//	if err != nil {
//		c.Error(err)
//		return "", err
//	}
//
//	if len(deviceShortList) == 0 {
//		return "", errors.New("没有设备需要控制")
//	}
//
//	//从redis中获取设备指令，当指令不存在时从涂鸦api去获取设备指令
//	values, err := db.GRedis.MGet(context.Background(), productId...).Result()
//	if err != nil {
//		c.Error(err)
//		return "", err
//	}
//
//	var cmdsMap = make(map[string][]*function)
//	for i, value := range values {
//		fmt.Println("-----redis--", value)
//
//		var fs []*function
//
//		if value == nil || value == "" || value == redis.Nil {
//
//			var cmdResp = &commandsResp{}
//			//从涂鸦api查询指令
//			err := tuyago.Get(c, fmt.Sprintf("/v1.0/devices/functions?device_ids=%s", productIdMap[productId[i]]), cmdResp)
//
//			if err != nil {
//				c.Error(err)
//				continue
//			}
//
//			if len(cmdResp.Result) == 0 || len(cmdResp.Result[0].Functions) == 0 {
//				continue
//			}
//
//			fsTmp := cmdResp.Result[0].Functions
//
//			fmt.Println("------2-1-32-123-", productId[i], productIdMap[productId[i]], x.MustMarshal2String(fsTmp))
//			//将指令存到redis
//			err = db.GRedis.Set(context.Background(), productId[i], x.MustMarshal2String(fsTmp), 0).Err()
//			if err != nil {
//				c.Error(err)
//				continue
//			}
//
//			fs = fsTmp
//
//		} else {
//			err = x.MustUnmarshal([]byte(value.(string)), fs)
//			if err != nil {
//				c.Error(err)
//				continue
//			}
//		}
//
//		cmdsMap[productId[i]] = fs
//	}
//
//	mcList := []llms.MessageContent{
//		{
//			Role:  llms.ChatMessageTypeHuman,
//			Parts: []llms.ContentPart{llms.TextPart("设备数据：" + x.MustMarshal2String(deviceShortList))},
//		},
//		{
//			Role:  llms.ChatMessageTypeHuman,
//			Parts: []llms.ContentPart{llms.TextPart("指令数据：" + x.MustMarshal2String(cmdsMap))},
//		},
//		{
//			Role:  llms.ChatMessageTypeHuman,
//			Parts: []llms.ContentPart{llms.TextPart(getFirstInput(c))},
//		},
//	}
//
//	//
//	result, err := langchaingoOpenAi.GenerateContent(
//		context.Background(),
//		mcList,
//		llms.WithTemperature(0.1),
//		llms.WithN(1),
//		llms.WithTopP(0.1),
//		llms.WithTools(openaiFunctionTools),
//	)
//
//	//var resultFullInfoDevice = make(map[string]*device)
//	////根据需要控制的设备信息，找到全量设备信息
//	//for i := range out.Result {
//	//	resultFullInfoDevice[out.Result[i].Id] = deviceFullInfoMap[out.Result[i].Id]
//	//}
//
//	//c.Debugf("need control device status now |data=%v |out=%v", x.MustMarshal2String(resultFullInfoDevice), out)
//	fmt.Println("-------2---2-3-2-32-3-", x.MustMarshal2String(result))
//	return "", nil
//}

// 过滤房间
func skipRoom(c *ava.Context, input string) ([]*room, error) {
	var homeId = getHomeId(c)

	//获取房间信息
	var roomResp = &roomInfo{}

	err := tuyago.Get(c, fmt.Sprintf("/v1.0/homes/%s/rooms", homeId), roomResp)

	if err != nil {
		c.Error(err)
		return nil, err
	}

	c.Debugf("roomInfo: %v", roomResp)

	//var isSkip = false
	////判断用户意图中是否有“非”关键字
	//for key := range skipWord {
	//	if strings.Contains(input, key) {
	//		isSkip = true
	//		break
	//	}
	//}

	//判断房间是否在用户的意图中
	var r = make([]*room, 0, 4)
	var tmpR = make([]*room, 0, len(roomResp.Result.Rooms))
	for i := range roomResp.Result.Rooms {
		tmpR = append(tmpR, roomResp.Result.Rooms[i])
		//如果用户的意图中有房间信息
		//if strings.Contains(input, roomResp.Result.Rooms[i].Name) && !isSkip {
		r = append(r, roomResp.Result.Rooms[i])
		//continue
		//}
	}

	if len(r) == 0 {
		r = nil
		r = tmpR
	}

	return r, nil
}

var defaultCategoryListKey = "TUYA_USER_HOME_CATEGORY_"
var defaultShortDevicesKey = "TUYA_USER_HOME_SHORT_DEVICES_"

// 获取用户设备种类和房间相关信息
func getCategoryList(c *ava.Context) (*CategoriesList, error) {
	result, err := db.GRedis.Get(context.Background(), defaultCategoryListKey+getHomeId(c)).Result()
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
		shortDeviceList, categoryList, err := allDevices(c, roomResp.Result.Rooms)
		if err != nil {
			c.Error(err)
			return nil, err
		}

		err = db.GRedis.Set(
			context.Background(),
			defaultCategoryListKey+getHomeId(c),
			x.MustMarshal2String(categoryList),
			time.Hour*2).Err()
		if err != nil {
			c.Error(err)
			return nil, err
		}

		err = db.GRedis.Set(
			context.Background(),
			defaultShortDevicesKey+getHomeId(c),
			x.MustMarshal2String(shortDeviceList),
			time.Hour*2).Err()
		if err != nil {
			c.Error(err)
			return nil, err
		}

		return categoryList, nil
	}

	var categoryList CategoriesList
	err = x.MustUnmarshal([]byte(result), &categoryList)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	return &categoryList, nil
}

func allDevices(c *ava.Context, r []*room) (map[string][]*shortDevice, *CategoriesList, error) {

	//返回的设备信息，key：房间名称
	var shortDevicesMap = make(map[string][]*shortDevice, 2)
	var categoriesList = new(CategoriesList)

	var mux = new(sync.Mutex)

	pool, err := ants.NewPool(len(r))
	if err != nil {
		c.Error(err)
		return nil, nil, err
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
			var categoryData []*CategoryData

			mux.Lock()

			var tmpDevice []*shortDevice

			for ii := range tmpResp.Result {
				tmpData := tmpResp.Result[ii]

				categoryName := getCategoryName(tmpData.Category)

				if categoryName == "" {
					c.Debugf("category name not exist |category=%v", tmpData.Category)
					continue
				}

				tmpDevice = append(tmpDevice, &shortDevice{
					Id:        tmpData.Id,
					Name:      tmpData.Name,
					Category:  tmpData.Category,
					ProductId: tmpData.ProductId,
				})

				if _, ok := skipCategoryMap[categoryName]; !ok {
					skipCategoryMap[categoryName] = true

					categoryData = append(categoryData, &CategoryData{
						CategoryName: categoryName,
						Category:     tmpData.Category,
					})

				}
			}

			if len(tmpDevice) == 0 {
				return
			}

			shortDevicesMap[tmpRoom.Name] = tmpDevice
			categories := &Categories{
				RoomName:     tmpRoom.Name,
				CategoryData: categoryData,
			}
			categoriesList.Categories = append(categoriesList.Categories, categories)
			mux.Unlock()
		})
	}

	err = pool.ReleaseTimeout(time.Second * 10)
	if err != nil {
		c.Error(err)
		return nil, nil, err
	}

	return shortDevicesMap, categoriesList, nil
}

// 获取需要控制的房间，设备类型
func getNeedToControl(c *ava.Context, toAI *CategoriesList, input string) (*CategoriesList, error) {
	mcList := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(fmt.Sprintf(filterDeviceCategory, x.MustMarshal2String(toAI)))},
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
		llms.WithTools(openaiCategoryTools),
	)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	if len(result.Choices) == 0 || result.Choices[0].FuncCall == nil || result.Choices[0].FuncCall.Arguments == "" {
		return nil, fmt.Errorf("ai no resp |result=%v", result)
	}

	c.Debugf("获取需要控制的设备，房间和设备类型信息 |result=%v", result.Choices[0].FuncCall.Arguments)

	var resp CategoriesList

	err = x.MustNativeUnmarshal([]byte(result.Choices[0].FuncCall.Arguments), &resp)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	if len(resp.Categories) == 0 {
		return nil, fmt.Errorf("ai no resp |resp=%v", resp)
	}

	c.Debugf("CategoryList: %v", x.MustMarshal2String(&resp))

	return &resp, nil
}

// 根据品类，将对应房间的，需要控制的品类的产品的所有信息发给ai
func controlDevices(c *ava.Context, categoryList *CategoriesList, input string) (*aiResp, error) {

	var devices map[string][]*shortDevice
	devicesStr, err := db.GRedis.Get(context.Background(), defaultShortDevicesKey+getHomeId(c)).Result()
	if err != nil || devicesStr == "" {
		c.Error(err)
		return nil, err
	}

	err = x.MustNativeUnmarshal([]byte(devicesStr), &devices)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	var deviceCommandList DeviceCommandList

	for i := range categoryList.Categories {
		var categories = categoryList.Categories[i]
		var roomName = categories.RoomName
		var tmpDeviceCommandDataArr = &DeviceCommandInfo{RoomName: roomName}

		for ii := range categories.CategoryData {
			var productIdMap = make(map[string]*DeviceCommandData, 10) //用来过滤重复的productId
			var category = categories.CategoryData[ii].Category
			var categoryName = categories.CategoryData[ii].CategoryName

			//找出当前房间该品类的设备
			devicesList, ok := devices[roomName]
			if !ok {
				continue
			}

			for iii := range devicesList {
				var device = devicesList[iii]

				//当类型一样的时候
				if device.Category == category {
					var tmpDeviceCommandData *DeviceCommandData

					var short = &shortDevice{
						Id:        device.Id,
						Name:      device.Name,
						Category:  device.Category,
						ProductId: device.ProductId,
					}
					if deviceCommandData, ok := productIdMap[device.ProductId]; !ok {
						f, err := getDeviceFunction(c, device.ProductId, device.Category, device.Id)
						if err != nil {
							c.Error(err)
							continue
						}

						if len(f) == 0 {
							continue
						}

						var s []*shortDevice
						s = append(s, short)

						tmpDeviceCommandData = new(DeviceCommandData)
						tmpDeviceCommandData.Commands = f
						tmpDeviceCommandData.CategoryName = categoryName
						tmpDeviceCommandData.ProductId = device.ProductId
						tmpDeviceCommandData.Devices = s

						productIdMap[device.ProductId] = tmpDeviceCommandData
					} else {
						deviceCommandData.Devices = append(deviceCommandData.Devices, short)
					}

					tmpDeviceCommandDataArr.DeviceCommandData = append(tmpDeviceCommandDataArr.DeviceCommandData, tmpDeviceCommandData)
				}
			}
		}

		deviceCommandList.Result = append(deviceCommandList.Result, tmpDeviceCommandDataArr)
	}

	//将数据发送给ai，让ai返回控制指令
	mcList := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(fmt.Sprintf(getDevicesCommand, x.MustMarshal2String(&deviceCommandList)))},
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
		llms.WithTools(openaiFunctionTools),
	)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	if len(result.Choices) == 0 || result.Choices[0].FuncCall == nil || result.Choices[0].FuncCall.Arguments == "" {
		return nil, fmt.Errorf("ai no resp |result=%v", result)
	}

	var resp aiResp

	err = x.MustNativeUnmarshal([]byte(result.Choices[0].FuncCall.Arguments), &resp)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	return &resp, nil

}

var defaultProductFunctionPrefix = "TUYA_PRODUCT_ID_FUNCTION_"

// 获取设备指令，如果不存在择用deviceId重新获取并写到redis
func getDeviceFunction(c *ava.Context, productID, category, deviceId string) ([]*function, error) {
	result, err := db.GRedis.Get(context.Background(), defaultProductFunctionPrefix+category+"_"+productID).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		c.Error(err)
		return nil, err
	}

	if result == "-" {
		return nil, fmt.Errorf("this product function is empty")
	}

	var resp []*function

	if result == "" || errors.Is(err, redis.Nil) {
		//从涂鸦获取设备指令
		var cmdResp = &commandsResp{}
		//从涂鸦api查询指令
		err = tuyago.Get(c, fmt.Sprintf("/v1.0/devices/functions?device_ids=%s", deviceId), cmdResp)

		if err != nil {
			c.Errorf("productId=%s |deviceId=%s |err=%v", productID, deviceId, err)
			return nil, err
		}

		if cmdResp.Success && len(cmdResp.Result) == 0 {
			err = db.GRedis.Set(context.Background(), defaultProductFunctionPrefix+category+"_"+productID, "-", 0).Err()
			if err != nil {
				ava.Error(err)
				return nil, err
			}

			c.Errorf("productId=%s |deviceId=%s |err=%v", productID, deviceId, err)
			return nil, fmt.Errorf("this productId no data |productId=%v", productID)
		}

		resp = cmdResp.Result[0].Functions
	} else {
		err = x.MustUnmarshal([]byte(result), &resp)
		if err != nil {
			c.Errorf("productId=%s |deviceId=%s |err=%v", productID, deviceId, err)
			return nil, err
		}

		if len(resp) == 0 {
			c.Errorf("productId=%s |deviceId=%s |err=%v", productID, deviceId, err)
			return nil, fmt.Errorf("this productId no functions |productId=%v", productID)
		}
	}

	err = db.GRedis.Set(context.Background(), defaultProductFunctionPrefix+category+"_"+productID, x.MustMarshal2String(resp), 0).Err()
	if err != nil {
		ava.Error(err)
		return nil, err
	}

	return resp, err
}

type DeviceCommandList struct {
	Result []*DeviceCommandInfo `json:"result"`
}

type DeviceCommandInfo struct {
	RoomName          string               `json:"room_name"`
	DeviceCommandData []*DeviceCommandData `json:"device_command_data"`
}

// 按照productId将产品指令归类发送给ai
type DeviceCommandData struct {
	ProductId    string         `json:"productId"`
	CategoryName string         `json:"categoryName"`
	Devices      []*shortDevice `json:"devices"`
	Commands     []*function    `json:"commands"`
}

// 用于ai返回，所有类型，ai不能直接识别[],[]必须有字段名称才能识别并返回
type CategoriesList struct {
	Categories []*Categories `json:"categories"`
}

// 单个房间里的设备类型
type Categories struct {
	RoomName     string          `json:"room_name"`
	CategoryData []*CategoryData `json:"category_data"`
}

type CategoryData struct {
	CategoryName string `json:"category_name"`
	Category     string `json:"category"`
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

// ai返回内容格式
type aiResp struct {
	Voice      string       `json:"voice"`
	Result     []aiRespData `json:"result"`
	ResultType int          `json:"result_type"`
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

type deviceResp struct {
	Result   *device `json:"result"`
	Position string  `json:"position"` //非接口返回字段，需要通过其他接口获取
	Success  bool    `json:"success"`
	T        int     `json:"t"`
	Tid      string  `json:"tid"`
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
	Id        string `json:"id"`
	Name      string `json:"name"`
	Category  string `json:"category"`
	ProductId string `json:"product_id"`
}

var openaiCategoryTools = []llms.Tool{
	{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "choose_category_list",
			Description: "根据我的意图和设备种类信息，筛选出即将要控制的设备。",
			Parameters: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"categories": {
						Type:        jsonschema.Array,
						Description: "房间及其设备种类信息",
						Items: &jsonschema.Definition{
							Type: jsonschema.Object,
							Properties: map[string]jsonschema.Definition{
								"room_name": {
									Type:        jsonschema.String,
									Description: "房间名称",
								},
								"category_data": {
									Type:        jsonschema.Array,
									Description: "设备种类数据",
									Items: &jsonschema.Definition{
										Type: jsonschema.Object,
										Properties: map[string]jsonschema.Definition{
											"category_name": {
												Type:        jsonschema.String,
												Description: "种类名称",
											},
											"category": {
												Type:        jsonschema.String,
												Description: "种类类型id",
											},
										},
										Required: []string{"category_name", "category"},
									},
								},
							},
							Required: []string{"room_name", "category_data"},
						},
					},
				},
				Required: []string{"categories"},
			},
		},
	},
}

var openaiFunctionTools = []llms.Tool{
	{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "get_device_command",
			Description: "获取设备的控制指令，如果是Boolean类型，values的值是true或false",
			Parameters: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"voice": {
						Type:        jsonschema.String,
						Description: "对应用户语音响应，用淘气的语气，包含对每个设备的控制详情",
					},
					"result": {
						Type:        jsonschema.Array,
						Description: "设备列表及其控制指令",
						Items: &jsonschema.Definition{
							Type: jsonschema.Object,
							Properties: map[string]jsonschema.Definition{
								"id": {
									Type:        jsonschema.String,
									Description: "设备ID",
								},
								"name": {
									Type:        jsonschema.String,
									Description: "设备名称",
								},
								"commands": {
									Type:        jsonschema.Array,
									Description: "当前设备需要执行的控制指令",
									Items: &jsonschema.Definition{
										Type: jsonschema.Object,
										Properties: map[string]jsonschema.Definition{
											"code": {
												Type:        jsonschema.String,
												Description: "控制指令类型",
											},
											"value": {
												Type:        jsonschema.String,
												Description: "控制指令内容值",
											},
										},
										Required: []string{"code", "value"},
									},
								},
							},
							Required: []string{"id", "name", "commands"},
						},
					},
				},
				Required: []string{"voice", "result"},
			},
		},
	},
}
