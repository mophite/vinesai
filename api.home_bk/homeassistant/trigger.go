package homeassistant

// 根据wifi的bssid判断用户在什么位置
var BSSID = map[string]map[string]string{
	"123": {
		"78:60:5B:85:3A:98": "客厅",
		"78:60:5b:85:3a:9a": "卧室",
	},
}

//定时器
//人体传感器
//温度
//湿度
//天气（室外气候，室外温度，风速）
//时间（扫地机器人工作）
