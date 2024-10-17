package langchain

type deviceResp struct {
	Result   *device `json:"result"`
	Position string  `json:"position"` //非接口返回字段，需要通过其他接口获取
	Success  bool    `json:"success"`
	T        int     `json:"t"`
	Tid      string  `json:"tid"`
}

// 用户设备信息
type device struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	Status    []*status `json:"status"`
	Category  string    `json:"category"`
	Online    bool      `json:"online"`
	ProductId string    `json:"product_id"`

	roomName string
}

type status struct {
	Code  string      `json:"code"`
	Value interface{} `json:"value"`
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
	Result   []*device `json:"result"`
	Position string    `json:"position"` //非接口返回字段，需要通过其他接口获取
	Success  bool      `json:"success"`
	T        int       `json:"t"`
	Tid      string    `json:"tid"`
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
	RoomID int64  `json:"room_id"`
}

type ShortSummaryDeviceInfo struct {
	Result []*ShortSummaryDevice `json:"result"`
}

type ShortSummaryDevice struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Position  string `json:"position"`
	ProductId string `json:"product_id"`
	Category  string `json:"category"`
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
	Content string   `json:"content"`
	Summary string   `json:"summary"`
	Devices []string `json:"devices"`
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
