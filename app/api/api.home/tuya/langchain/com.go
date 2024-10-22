package langchain

// 设备详情
type deviceResp struct {
	Result   *mgoDocDevice `json:"result"`
	Position string        `json:"position"` //非接口返回字段，需要通过其他接口获取
	Success  bool          `json:"success"`
	T        int           `json:"t"`
	Tid      string        `json:"tid"`
}

var mgoCollectionNameDevice = "tuya_device"
var mgoCollectionNameCodes = "tuya_codes"

// 用户设备信息
type mgoDocDevice struct {
	ID          string   `json:"id" bson:"_id"`
	ActiveTime  int      `json:"active_time"`
	BizType     int      `json:"biz_type"`
	Category    string   `json:"category"`
	CreateTime  int      `json:"create_time"`
	Icon        string   `json:"icon"`
	IP          string   `json:"ip"`
	Lat         string   `json:"lat"`
	LocalKey    string   `json:"local_key"`
	Lon         string   `json:"lon"`
	Model       string   `json:"model"`
	Name        string   `json:"name"`
	Online      bool     `json:"online"`
	OwnerID     string   `json:"owner_id"`
	ProductID   string   `json:"product_id"`
	ProductName string   `json:"product_name"`
	Status      []status `json:"status"`
	Sub         bool     `json:"sub"`
	TimeZone    string   `json:"time_zone"`
	UID         string   `json:"uid"`
	UpdateTime  int      `json:"update_time"`
	UUID        string   `json:"uuid"`

	//以下非接口直接返回数据
	RoomName     string `json:"room_name"`
	HomeId       string `json:"home_id"`
	RoomId       int    `json:"room_id"`
	CategoryName string `json:"category_name"`
	//HomeName string `json:"home_name"`
}

// 批量获取指令集
type commandsResp struct {
	Result  []*command `json:"result"`
	Success bool       `json:"success"`
	T       int        `json:"t"`
	Tid     string     `json:"tid"`
}

type command struct {
	Devices   []string    `json:"devices"`
	Functions []*function `json:"functions"`
}

type function struct {
	Code   string      `json:"code"`
	Desc   string      `json:"desc"`
	Name   string      `json:"name"`
	Type   string      `json:"type"`
	Values interface{} `json:"values"`
}

// 获取用户的设备列表
type deviceListResp struct {
	Result   []*mgoDocDevice `json:"result"`
	Position string          `json:"position"` //非接口返回字段，需要通过其他接口获取
	Success  bool            `json:"success"`
	T        int             `json:"t"`
	Tid      string          `json:"tid"`
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

type room struct {
	Name   string `json:"name"`
	RoomID int    `json:"room_id"`
}

type ShortDevice struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type summaryCommandsResp struct {
	FailureMsg string `json:"failure_msg"`
	SuccessMsg string `json:"success_msg"`
	Result     []struct {
		Code  string      `json:"code"`
		Value interface{} `json:"value"`
	} `json:"result"`
}

// 单个房间里的设备类型
type SummaryCategories struct {
	CategoryData []*summaryCategoryData `json:"category_data"`
}

type summaryCategoryData struct {
	CategoryName string `json:"category_name"`
	Category     string `json:"category"`
}

// 一次多个意图
type summaries struct {
	FailureMsg string         `json:"failure_msg"`
	Result     []*summaryData `json:"result"`
}

type summaryData struct {
	Content string `json:"content"`
	Summary string `json:"summary"`
	Devices string `json:"devices"` //逗号分割
}

type queryOnlineResp struct {
	Content string `json:"content"`
}

type queryOnlineOrOfflineData struct {
	Name         string
	CategoryName string
	RoomName     string
}

type queryDevicesData struct {
	Name         string
	CategoryName string
	//RoomName     string
	Status []status
}

type status struct {
	Code  string      `json:"code"`
	Value interface{} `json:"value"`
}

// 查询支持场景的设备列表
type supportSceneDevices struct {
	Result []struct {
		DeviceID string `json:"device_id"`
		Name     string `json:"name"`
	} `json:"result"`
	Success bool   `json:"success"`
	T       int64  `json:"t"`
	Tid     string `json:"tid"`
}

// 获取家庭设备支持的场景动作，触发事件
type homeCodeAndStatusResp struct {
	Result  []homeFunctionAndStatus `json:"result"`
	Success bool                    `json:"success"`
	T       int                     `json:"t"`
	Tid     string                  `json:"tid"`
}

type homeFunctionAndStatus struct {
	DeviceId  string        `json:"device_id" bson:"_id"`
	Functions []interface{} `json:"functions"`
	Status    []interface{} `json:"status"`
	//非接口返回字段
	Name     string `json:"name"`
	HomeId   string `json:"home_id"`
	RoomName string `json:"room_name"`
}

type homeFunction struct {
	DeviceId  string        `json:"device_id" bson:"_id"`
	Functions []interface{} `json:"functions"`
}

type createScene struct {
	Name       string    `json:"name"`
	Background string    `json:"background"`
	Actions    []actions `json:"actions"`
	Content    string    `json:"content"`
}

type actions struct {
	ExecutorProperty interface{} `json:"executor_property"`
	ActionExecutor   string      `json:"action_executor"`
	EntityID         string      `json:"entity_id"`
}

func getCategoryName(categoryId string) string {
	return testcategoryName[categoryId]
}

// 设备类型对应中文名称
// todo区分主动控制设备，和传感器等被动设备
var testcategoryName = map[string]string{
	"dj": "灯具，包括灯带，筒灯，射灯，轨道灯,双色灯，多色灯等",
	"kg": "开关",
	"pc": "插排",
}

var categoryName = map[string]string{
	"dj":  "灯具，包括灯带，筒灯，射灯，轨道灯,双色灯，多色灯等",
	"xdd": "吸顶灯",
	"fwd": "氛围灯",
	//"dc":          "灯串",
	//"dd":          "灯带",
	"gyd": "感应灯",
	//"fsd": "风扇灯",
	//"tyndj":       "太阳能灯具",
	"tgq": "调光器",
	//"ykq":         "遥控器",
	//"sxd":         "射灯",
	"kg":   "开关",
	"cz":   "插座",
	"pc":   "插排",
	"cjkg": "场景开关",
	//"ckqdkg": "插卡取电开关",
	"clkg": "窗帘开关",
	//"ckmkzq": "车库门控制器",
	"tgkg": "调光开关",
	//"fskg":   "风扇开关",
	//"wxkg":        "无线开关",
	//"rs":          "热水器",
	//"xfj":         "新风机",
	//"bx":          "冰箱",
	//"yg":          "浴缸",
	//"xy":          "洗衣机",
	//"kt":          "空调",
	"ktkzq": "空调控制器",
	//"bgl":         "壁挂炉",
	//"sd":          "扫地机器人",
	//"qn":          "取暖器",
	//"kj":          "空气净化器",
	"lyj": "晾衣架",
	"xxj": "香薰机",
	"cl":  "窗帘",
	"mc":  "门窗控制器",
	"wk":  "温控器",
	"yb":  "浴霸",
	//"ggq":         "灌溉器",
	//"jsq":         "加湿器",
	//"cs":          "除湿机",
	//"fs":          "风扇",
	//"js":          "净水器",
	//"dr":          "电热毯",
	//"cwtswsq":     "宠物弹射喂食器",
	//"cwwqfsq":     "宠物网球发射器",
	//"ntq":         "暖通器",
	//"cwwsq":       "宠物喂食器",
	//"cwysj":       "宠物饮水机",
	//"sf":          "沙发",
	//"dbl":         "电壁炉",
	//"tnq":         "智能调奶器",
	//"msp":         "猫砂盆",
	//"mjj":         "毛巾架",
	//"sz":          "植物生长机",
	//"bh":          "智能电茶壶",
	//"mb":          "面包机",
	//"kfj":         "咖啡机",
	//"nnq":         "暖奶器",
	//"cn":          "冲奶机",
	//"mzj":         "慢煮机",
	//"mg":          "米柜",
	//"dcl":         "电磁炉",
	//"kqzg":        "空气炸锅",
	//"znfh":        "智能饭盒",
	"wg2":   "网关",
	"mal":   "报警主机",
	"sp":    "智能摄像机",
	"sgbj":  "声光报警-传感器",
	"rqbj":  "燃气报警-传感器",
	"ywbj":  "烟雾报警-传感器",
	"wsdcg": "温湿度-传感器",
	"mcs":   "门磁-传感器",
	"zd":    "震动-传感器",
	"sj":    "水浸-传感器",
	"ldcg":  "亮度-传感器",
	"ylcg":  "压力-传感器",
	"sos":   "紧急按钮",
	"pm2.5": "PM2.5-传感器",
	"cobj":  "CO-传感器",
	"co2bj": "CO2-传感器",
	"dgnbj": "多功能传感器",
	"jwbj":  "甲烷报警-传感器",
	"pir":   "人体运动传感器",
	"hps":   "人体存在传感",
	"ms":    "家用门锁",
	//"bxx":         "保险箱",
	"gyms":        "公寓门锁",
	"jtmspro":     "家用门锁 PRO",
	"hotelms":     "酒店门锁",
	"ms_category": "门锁配件",
	"jtmsbh":      "家用门锁保活",
	"mk":          "门控",
	"videolock":   "视频锁",
	"photolock":   "音视频锁",
	"hjjcy":       "环境检测仪",
	//"amy":         "按摩椅",
	//"liliao":      "理疗产品",
	//"ts":          "跳绳",
	//"tzc1":        "体脂秤",
	//"sb":          "手表/手环",
	//"znyh":        "智能药盒",
	//"zndb":        "智能电表",
	//"znsb":        "智能水表",
	"dlq": "断路器",
	//"ds":          "电视",
	//"tyy":         "投影仪",
	"tracker": "定位器",
}
