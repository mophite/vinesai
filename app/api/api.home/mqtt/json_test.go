package mqtt

import (
	"encoding/json"
	"fmt"
	"testing"
)

type Device2 struct {
	ID         uint   `gorm:"column:id;primary_key;AUTO_INCREMENT" json:"id"`
	DeviceType string `gorm:"column:device_type" json:"device_type"`
	DeviceName string `gorm:"column:device_name" json:"device_name"`
	DeviceId   string `gorm:"column:device_id" json:"device_id"`
	UserId     string `gorm:"column:user_id;NOT NULL" json:"user_id"`
	Action     string `gorm:"column:action" json:"action"`
	Timestamp  int    `gorm:"column:timestamp" json:"timestamp"`
	Data       string `gorm:"column:data" json:"data"`
	CreatedAt  string `gorm:"column:created_at;<-:false;default:null" json:"created_at"` // 数据入库时间
	UpdatedAt  string `gorm:"column:updated_at;<-:false;default:null" json:"updated_at"` // 数据修改时间
}

func TestUnamsh(t *testing.T) {
	var result = struct {
		Result  []*Device2 `json:"result"`
		Message string     `json:"message"`
	}{}

	var s = `{"result":[{"id":100002,"device_id":"34334546","action":"turn_on","data":{"temperature":25.0,"switch":"turn_on"}}],"message":"好的，主人。已经将中央空调打开，温度调整为25摄氏度。"}`
	json.Unmarshal([]byte(s), &result)
	c, _ := json.Marshal(&result)
	fmt.Println(string(c))

	var b []string
	b = append(b, "1", "2")
	fmt.Println(b)
}
