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


