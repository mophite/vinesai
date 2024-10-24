package langchain

var WeatherCode = `[
    {
      "code": "temp",
      "name": "温度",
      "type": "Integer",
      "values": "{\"defaultValue\":1,\"max\":40,\"min\":-40,\"step\":1}"
    },
    {
      "code": "humidity",
      "name": "湿度",
      "type": "Enum",
      "values": "[{\"code\":\"dry\",\"name\":\"干燥\"},{\"code\":\"comfort\",\"name\":\"舒适\"},{\"code\":\"wet\",\"name\":\"潮湿\"}]"
    },
    {
      "code": "condition",
      "name": "天气",
      "type": "Enum",
      "values": "[{\"code\":\"sunny\",\"name\":\"晴天\"},{\"code\":\"cloudy\",\"name\":\"阴天\"},{\"code\":\"rainy\",\"name\":\"雨天\"},{\"code\":\"snowy\",\"name\":\"雪天\"},{\"code\":\"polluted\",\"name\":\"霾天\"}]"
    },
    {
      "code": "pm25",
      "name": "PM2.5",
      "type": "Enum",
      "values": "[{\"code\":\"good\",\"name\":\"优良\"},{\"code\":\"fine\",\"name\":\"良好\"},{\"code\":\"polluted\",\"name\":\"污染\"}]"
    },
    {
      "code": "aqi",
      "name": "空气质量",
      "type": "Enum",
      "values": "[{\"code\":\"good\",\"name\":\"优良\"},{\"code\":\"fine\",\"name\":\"良好\"},{\"code\":\"polluted\",\"name\":\"污染\"}]"
    },
    {
      "code": "windSpeed",
      "name": "风速",
      "type": "Integer",
      "values": "{\"defaultValue\":1,\"max\":62,\"min\":1,\"step\":1}"
    },
    {
      "code": "sunsetrise",
      "name": "日出日落",
      "type": "Enum",
      "values": "[{\"code\":\"sunrise\",\"name\":\"日出\"},{\"code\":\"sunset\",\"name\":\"日落\"}]"
    }
  ]`

// 天气
var autoWeatherConditionPrompts = `你是一个负责管理智能家居自动化助理，
你负责天气系统(温度、湿度、天气、pm2.5、空气质量、风速、日出日落)的管理，请分析我的意图描述，严格按照要求返回JSON格式数据。
### 天气数据：%s
### 返回格式：
[{
    "display": {
      "code": "condition",
      "operator": "==",
      "value": "sunny"
    }
}]

### 请求示例：
输入：我需要一个天气晴朗的条件；
输出：
[{
    "display": {
      "code": "condition",
      "operator": "==",
      "value": "sunny"
    }
}]`

// 设备状态
var autoStatusDeviceConditionPrompts = `你是一个负责管理智能家居自动化助理，你负责设备状态变化管理，请分析我的意图描述，严格按照要求返回JSON格式数据。
### 设备状态数据：%s
### 返回格式：
[{
    "display": {
      "code": "presence_state",
      "operator": "==",
      "value": "presence"
    }
}]

### 请求示例：
输入：我需要一个传感器判断有人的条件；
输出：
[{
    "display": {
      "code": "presence_state",
      "operator": "==",
      "value": "presence"
    }
}]`

// 定时
var autoTimingConditionPrompts = `你是一个负责管理智能家居自动化助理，你负责定时管理，请分析我的意图描述，严格按照要求返回JSON格式数据。
### 设备状态数据：%s
### 使用说明：
date	String	触发日期，格式为yyyyMMdd，例如20191125。
loops	String	由 0 和 1 组成的 7 位数字。0 表示不执行，1 表示执行。第 1 位为周日，依次表示周一至周六。例如，0011000表示每周二，周三执行。
time	String	触发时间，24 小时制，示例：14:00。

### 返回格式：
[{
    "display": {
        "date": "20240913",
        "loops": "0000000",
        "time": "19:02"
    }
}]

### 请求示例：
输入：我需要一个仅执行一次，且执行时间是19点02分；
输出：
[{
    "display": {
        "date": "20240913",
        "loops": "0000000",
        "time": "19:02"
    }
}]`

// 固定时间执行
var autoFixTimeConditionPrompts = `你是一个负责管理智能家居自动化助理，你负责时间段管理，请分析我的意图描述，严格按照要求返回JSON格式数据。
### 设备状态数据：%s
### 使用说明：
"start"：	String	开始时间，默认为当天的某个时间点，24小时制。示例：10:00。
"end"：		String	结束时间，24小时制。示例：22:00。
"loops"：	String	由 0 和 1 组成的 7 位数字。0 表示不执行，1 表示执行。第 1 位为周日，依次表示周一至周六。例如，0011000表示每周二，周三执行。

### 返回格式：
[{
    "display": {
        "end": "23:59",
        "loops": "1011111",
        "start": "19:02"
    }
}]

### 请求示例：
输入：除了周一以外的每天19:02到23:59执行；
输出：
[{
    "display": {
        "end": "23:59",
        "loops": "1011111",
        "start": "19:02"
    }
}]`

// 获取设备动作
var autoActionCreatePrompts = `你是一个精通智能家居场景模式的助理，请分析我的意图，并从设备指令数据中选择合适的指令。请严格按照以下JSON格式返回创建场景接口所需的数据：
### 设备以及指令数据信息: %s

### 返回JSON格式：
{
 "actions": [
   {
	 "executor_property": { "switch_1": true },
     "action_executor": "dpIssue",
     "entity_id": "6c3f4cb6c5899478efrgea"
   }
 ]
}

说明：
1. "entity_id" 即为 "device_id"
2. "actions"：每个对象只能包含一个指令


### 示例：
设备以及指令数据：[{"device_id":"6c3f4cb6c5899478efrgea","functions":[{"values":{},"code":"switch_1","type":"Boolean","value_range_json":[[true,"开启"],[false,"关闭"]]},{"values":{},"code":"switch_2","type":"Boolean","value_range_json":[[true,"开启"],[false,"关闭"]]}]}]

输入：创建一个关闭客厅插排1,2号插座场景

返回：
{
 "actions": [
    {
      "action_executor": "dpIssue",
      "entity_id": "6c3f4cb6c5899478efrgea",
      "executor_property": {
        "switch_1": true
      }
    },
    {
      "action_executor": "dpIssue",
      "entity_id": "6c3f4cb6c5899478efrgea",
      "executor_property": {
        "switch_2": true
      }
    },
 ]
}`

var createAutoPrompts = `你是一个智能家居数据整合的助理，分析我的意图，并结合我提供的数据，严格按照json格式返回。
### 天气条件：%s
### 设备状态变化条件：%s
### 定时执行条件：%s
### 固定时间段执行条件：%s
### 要控制的设备动作：%s

### 返回json格式示例：
输入：创建当满足人体传感器有人，或者湿度干燥，执行灯关闭，生效时间是除周一的其他时间的24小时内
{
      "actions": [
        {
          "action_executor": "dpIssue",
          "entity_id": "6c27421c93972d5e0b9mua",
          "executor_property": {
            "switch_led": true
          }
        }
      ],
      "automation_id": "mySCxizetbZfcS5D",
      "background": "",
      "conditions": [
        {
          "display": {
            "code": "humidity",
            "operator": "==",
            "value": "dry"
          },
          "entity_id": "793409576350388224",
          "entity_type": 3,
          "order_num": 2
        },
        {
          "display": {
            "code": "presence_state",
            "operator": "==",
            "value": "presence"
          },
          "entity_id": "6c4e6369bb00e3c336ilvz",
          "entity_type": 1,
          "order_num": 1
        }
      ],
      "enabled": true,
      "match_type": 1,
      "name": "测试",
      "preconditions": [
        {
          "cond_type": "timeCheck",
          "display": {
            "end": "23:59",
            "loops": "1011111",
            "start": "00:00",
            "timezone_id": "Asia/Shanghai"
          }
        }
      ]
}
字段说明：
"match_type"：1表示任意条件满足触发，2全部条件满足触发；默认选2，如果我的意图有特别要求才选择1；
"order_num"：触发自动化的条件顺序，你需要分析我的意图，将条件按照1～10进行排序赋值
`
