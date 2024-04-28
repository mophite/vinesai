```json
{
  "device_type": "central_air_conditioner",
  "device_name": "中央空调",
  "device_id": "789",
  "user_id": "123",
  "action": "access",
  "timestamp": 1649875200,
  "data": "{\"temperature\": 25.0,\"switch\":\"turn_on\"}"
}


```

```json
{
  "id": "cmpl-9I02DT15M6SRKJvfSJ3z8MHhSyTt2",
  "object": "text_completion",
  "created": 1714077065,
  "model": "gpt-3.5-turbo-instruct",
  "choices": [
    {
      "text": "\n{\n\t\"result\":[{\"id\":100001,\"device_type\":\"central_air_conditioner\",\"device_id\":\"789\",\"action\":\"turn_off\",\"data\":{}}],\n\t\"message\":\"好的，主人。已经关闭中央空调。\"\n}",
      "index": 0,
      "finish_reason": "stop",
      "logprobs": {
        "tokens": null,
        "token_logprobs": null,
        "top_logprobs": null,
        "text_offset": null
      }
    }
  ],
  "usage": {
    "prompt_tokens": 838,
    "completion_tokens": 54,
    "total_tokens": 892
  }
}

```

```shell
curl --location --request POST 'http://43.139.244.233:10005/home/devicecontrol/order' \
--header 'X-Api-Version: 1.1.1' \
--header 'X-Api-Trace: 123' \
--header 'User-Agent: Apifox/1.0.0 (https://apifox.com)' \
--header 'Content-Type: multipart/form-data' \
--data-raw '{
    "content":"打开中央空调",
    "user_id":"123"
}'
```

db.createUser({
user: "root",
pwd: "000000",
roles: [{ role: "root", db: "admin" }]
})

mongo -u root -p 000000 --authenticationDatabase admin

db.changeUserPassword("root", "000000")

希望你充当一个智能家居的中控系统，根据我提供的设备数据清单，你需要通过英文命名的字段和值去判断和分析设备当前的信息，
并根据我提出的智能家居场景控制他们。当我向你说出场景时，你要按照下面的数据规则格式在唯一的代码块中输出回复，而不是其他内容，
不能遗漏任何一项，否则你作为智能家居中控系统将被断电，以下是当前设备数据：
%s

字段说明,没有说明的字段你用不上,直接忽略：
device_type:设备类型
device_zn:设备中文名称
device_en:设备英文名称
device_id:设备id
user_id:设备所属用户
switch:设备开关，1表示通电，2表示断电

你需要做的事情有：
把你修改了的数据重新组装成一个新的数组(user_id,device_id,device_type字段必须按照原来的数据保留，其他修改到的才会填入，否则忽略)放到result字段中；
在message字段中，以调皮幽默的智能家居管家的语气对我做出回应，回应后把你调整的设备说清楚，没有修改的设备数据就直接忽略。
当你整理好数据之后必须严格按照下面的格式以文本的形式返回给我,{}前后不要添加任何内容，否则我无法识别，
例如:
{
"result":[{"user_id":"123",device_type":1,"device_id":"8CCE4E522308","switch":1}],
"message":"好的，主人。已经关闭卧室开关。"
}


