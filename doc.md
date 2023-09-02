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
    "msg": "指令获取成功"
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