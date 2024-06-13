## 请求头

```http header
    1.X-Api-Trace:日志追踪id
    2.X-Api-Version:使用的api版本
    3.Content-Type:application/json
    4.Authorization: Bearer token (有空格)
```

## 鉴权
- 地址:/hub/v1/token


- 请求
```json
{
    "grant_type":"client_credentials",
    "client_id":"498715320649678",
    "client_secret":"JYNFA9OHGQBL5IU62ZWMXKPS1TRD73VC"
} 
```

- 响应
```json
{
    "code": 200,
    "msg": "获取token成功",
    "data": {
        "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJUaW1lc3RhbXAiOiIyMDIzLTA4LTIxVDE0OjMwOjA3KzA4OjAwIiwiaXNzIjoidmluZXNhaSIsInN1YiI6Im9hdXRoMi4w5o6I5p2DIiwiYXVkIjpbImFwaS5vYXV0aDIuMCJdLCJleHAiOjQ4MDI5OTk0MDcsIm5iZiI6MTY5MjU5OTQwNywiaWF0IjoxNjkyNTk5NDA3LCJqdGkiOiJiMzA1NTBmZi02ZjM0LTQzOGMtYWYzOS1kMTY5OWFiZDkyNDUifQ.M9n11DV2XLKr_k5aKTQ4L4n-Vv8-O4iGwCywVumFU5U",
        "expires_in": 4802999407,
        "token_type": "Bearer"
    }
}
```

## 设备发现

- 请求
```json
{}
```

- 响应
```json
{
    "code": 200,
    "msg": "成功",
    "data": {
        "endpoints": [
            {
                "endpointId": "11",
                "friendlyName": "test",
                "description": "描述",
                "manufacturerName": "测试"
            }
        ]
    }
}
```

## 设备控制

- 请求
```json
{
    "controlAction":"关闭"
}
```

- 响应
```json
{
    "code": 200,
    "msg": x.StatusOk
}
```

## 设备状态

- 请求
```json
{}
```

- 响应
```json
{
    "code": 200,
    "msg": "成功",
    "devices": [
        {
            "deviceId": "123",
            "deviceAttributes": [
                {
                    "name": "测试",
                    "value": "测试"
                }
            ]
        }
    ]
}
```

## 设备上报

- 请求
```json
{
    "deviceAttributes": [
        {
            "deviceId": "123",
            "name": "客厅灯",
            "value": "123"
        }
    ]
}
```

- 响应
```json
{
    "code": 200,
    "msg": "成功",
    "data": [
        {
            "deviceId": "123",
            "status": "off",
            "message": "设备关闭中"
        }
    ]
}
```

```shell
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJkOTQyYzVlYWI1ZWI0OWVhOWQxNzM2MjhhMDg0YWNlMyIsImlhdCI6MTcxNjM5MzU0OCwiZXhwIjoyMDMxNzUzNTQ4fQ.qqpUAAALF4-DQOsPT6sTuEXaZX8REjJkGM5rQF2bghY" -H "Content-Type: application/json"  http://43.139.244.233:8123/api/services


curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJkOTQyYzVlYWI1ZWI0OWVhOWQxNzM2MjhhMDg0YWNlMyIsImlhdCI6MTcxNjM5MzU0OCwiZXhwIjoyMDMxNzUzNTQ4fQ.qqpUAAALF4-DQOsPT6sTuEXaZX8REjJkGM5rQF2bghY" -H "Content-Type: application/json"  -d '{"device_id": "85ae07efb54d69fb9d8a14ac25131b19"}'  http://43.139.244.233:8123/api/services/switch/toggle 
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJkOTQyYzVlYWI1ZWI0OWVhOWQxNzM2MjhhMDg0YWNlMyIsImlhdCI6MTcxNjM5MzU0OCwiZXhwIjoyMDMxNzUzNTQ4fQ.qqpUAAALF4-DQOsPT6sTuEXaZX8REjJkGM5rQF2bghY" -H "Content-Type: application/json"  -d '{"device_id": "30d1e889abd193b6b0c36337877510fc"}'  http://43.139.244.233:8123/api/services/switch/toggle 
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJkOTQyYzVlYWI1ZWI0OWVhOWQxNzM2MjhhMDg0YWNlMyIsImlhdCI6MTcxNjM5MzU0OCwiZXhwIjoyMDMxNzUzNTQ4fQ.qqpUAAALF4-DQOsPT6sTuEXaZX8REjJkGM5rQF2bghY" -H "Content-Type: application/json"   http://43.139.244.233:8123/api/states

curl -X POST -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJkOTQyYzVlYWI1ZWI0OWVhOWQxNzM2MjhhMDg0YWNlMyIsImlhdCI6MTcxNjM5MzU0OCwiZXhwIjoyMDMxNzUzNTQ4fQ.qqpUAAALF4-DQOsPT6sTuEXaZX8REjJkGM5rQF2bghY" -H "Content-Type: application/json"  -d '{"template": "{% set area_name = \"wo_shi\" %} {% set devices = area_devices(area_name) %} {% for device in devices %} {{ device.name }}: {{ device.entity_id }}<br> {% endfor %}"}'  http://43.139.244.233:8123/api/template
curl -X POST -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJkOTQyYzVlYWI1ZWI0OWVhOWQxNzM2MjhhMDg0YWNlMyIsImlhdCI6MTcxNjM5MzU0OCwiZXhwIjoyMDMxNzUzNTQ4fQ.qqpUAAALF4-DQOsPT6sTuEXaZX8REjJkGM5rQF2bghY" -H "Content-Type: application/json"  -d '{"template": "{{ area_id('wo_shi') }}"}'  http://43.139.244.233:8123/api/template


curl -X POST -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJkOTQyYzVlYWI1ZWI0OWVhOWQxNzM2MjhhMDg0YWNlMyIsImlhdCI6MTcxNjM5MzU0OCwiZXhwIjoyMDMxNzUzNTQ4fQ.qqpUAAALF4-DQOsPT6sTuEXaZX8REjJkGM5rQF2bghY" -H "Content-Type: application/json" -d '{"template": "{{ area_entities('\''wo_shi'\'') }}"}' http://43.139.244.233:8123/api/template


curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJkOTQyYzVlYWI1ZWI0OWVhOWQxNzM2MjhhMDg0YWNlMyIsImlhdCI6MTcxNjM5MzU0OCwiZXhwIjoyMDMxNzUzNTQ4fQ.qqpUAAALF4-DQOsPT6sTuEXaZX8REjJkGM5rQF2bghY" -H "Content-Type: application/json"  -d '{ "rgb_color": [255, 0, 0],"entity_id":"8de7933bfbabc360d650644840f5b600"}'  http://43.139.244.233:8123/api/services/light/turn_on 

curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJkOTQyYzVlYWI1ZWI0OWVhOWQxNzM2MjhhMDg0YWNlMyIsImlhdCI6MTcxNjM5MzU0OCwiZXhwIjoyMDMxNzUzNTQ4fQ.qqpUAAALF4-DQOsPT6sTuEXaZX8REjJkGM5rQF2bghY" \
-H "Content-Type: application/json" \
-d '{ "device_id": "bb5563b19238dc8689c90f12ad0608e4", "rgb_color": [255, 255, 0] }' \
http://127.0.0.1:8123/api/services/light/turn_on

curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJkOTQyYzVlYWI1ZWI0OWVhOWQxNzM2MjhhMDg0YWNlMyIsImlhdCI6MTcxNjM5MzU0OCwiZXhwIjoyMDMxNzUzNTQ4fQ.qqpUAAALF4-DQOsPT6sTuEXaZX8REjJkGM5rQF2bghY" \
-H "Content-Type: application/json" \
-d '{ "entity_id": "light.smart_led_strip", "rgb_color": [255, 0, 0] }' \
http://43.139.244.233:8123/api/services/light/turn_on
```

http://43.139.244.233:8123/auth/authorize?response_type=code&client_id=http://43.139.244.233:8123/&redirect_uri=http://43.139.244.233:8123/auth/callback

curl -X POST -H "Content-Type: application/x-www-form-urlencoded" \
-d "grant_type=authorization_code" \
-d "code=fda4623f714948629a17896053859e60" \
-d "client_id=http:/43.139.244.233:8123/" \
-d "redirect_uri=http:/43.139.244.233:8123/auth/callback" \
http:/43.139.244.233:8123/auth/token


curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJkOTQyYzVlYWI1ZWI0OWVhOWQxNzM2MjhhMDg0YWNlMyIsImlhdCI6MTcxNjM5MzU0OCwiZXhwIjoyMDMxNzUzNTQ4fQ.qqpUAAALF4-DQOsPT6sTuEXaZX8REjJkGM5rQF2bghY" \
-H "Content-Type: application/json" \
-d '{"entity_id":"light.smart_led_strip_2","rgb_color":[255,255,0]}' \
http://43.139.244.233:8123/api/services/light/turn_on

curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJkOTQyYzVlYWI1ZWI0OWVhOWQxNzM2MjhhMDg0YWNlMyIsImlhdCI6MTcxNjM5MzU0OCwiZXhwIjoyMDMxNzUzNTQ4fQ.qqpUAAALF4-DQOsPT6sTuEXaZX8REjJkGM5rQF2bghY" -H "Content-Type: application/json" -d '{"entity_id":"light.smart_led_strip_2","rgb_color":[255,255,0]}' http://localhost:8123/api/services/light/turn_on

curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJkOTQyYzVlYWI1ZWI0OWVhOWQxNzM2MjhhMDg0YWNlMyIsImlhdCI6MTcxNjM5MzU0OCwiZXhwIjoyMDMxNzUzNTQ4fQ.qqpUAAALF4-DQOsPT6sTuEXaZX8REjJkGM5rQF2bghY" \
-H "Content-Type: application/json" \
http://43.139.244.233:8123/api/config

curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJkOTQyYzVlYWI1ZWI0OWVhOWQxNzM2MjhhMDg0YWNlMyIsImlhdCI6MTcxNjM5MzU0OCwiZXhwIjoyMDMxNzUzNTQ4fQ.qqpUAAALF4-DQOsPT6sTuEXaZX8REjJkGM5rQF2bghY" \
-H "Content-Type: application/json" \
-d '{"entity_id":"remote.xiaomi_l05c_7e78_wifispeaker","message":"吃饭了吗"}' \
http://43.139.244.233:8123/api/services/tts/speak

curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJkOTQyYzVlYWI1ZWI0OWVhOWQxNzM2MjhhMDg0YWNlMyIsImlhdCI6MTcxNjM5MzU0OCwiZXhwIjoyMDMxNzUzNTQ4fQ.qqpUAAALF4-DQOsPT6sTuEXaZX8REjJkGM5rQF2bghY" \
-H "Content-Type: application/json" \
-d '{"entity_id"text.xiaomi_l05c_7e78_play_textxt","value":"你好"}' \
http://43.139.244.233:8123/api/services/text/set_value

curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJkOTQyYzVlYWI1ZWI0OWVhOWQxNzM2MjhhMDg0YWNlMyIsImlhdCI6MTcxNjM5MzU0OCwiZXhwIjoyMDMxNzUzNTQ4fQ.qqpUAAALF4-DQOsPT6sTuEXaZX8REjJkGM5rQF2bghY" \
-H "Content-Type: application/json" \
-d '{"entity_id"text.xiaomi_l05c_7e78_execute_text_directive","value":"打开空调"}' \
http://43.139.244.233:8123/api/services/text/set_value

curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJkOTQyYzVlYWI1ZWI0OWVhOWQxNzM2MjhhMDg0YWNlMyIsImlhdCI6MTcxNjM5MzU0OCwiZXhwIjoyMDMxNzUzNTQ4fQ.qqpUAAALF4-DQOsPT6sTuEXaZX8REjJkGM5rQF2bghY" \
-H "Content-Type: application/json" \
-d '{"entity_id":"media_player.xiaomi_l05c_7e78"}' \
http://43.139.244.233:8123/api/intent/handle


curl -X POST 'http://223.72.19.182:8000/v1/workflows/run' \
--header 'Authorization: Bearer app-gmhMQjOwBtsyYMfvlcQDx0E0' \
--header 'Content-Type: application/json' \
--data-raw '{
"inputs": {"query":"your name?"},
"response_mode": "blocking",
"user": "abc-123"
}'
