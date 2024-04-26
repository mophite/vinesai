package mqtt

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestUnamsh(t *testing.T) {
	var result = struct {
		Result  interface{} `json:"result"`
		Message string      `json:"message"`
	}{}

	var s = `{
	"result":[{"id":100001,"device_type":"central_air_conditioner","device_id":"789","action":"turn_off"}],
	"message":"好的，主人。已经关闭中央空调。"
}`

	json.Unmarshal([]byte(s), &result)
	fmt.Println(result)

	var b []string
	b = append(b, "1", "2")
	fmt.Println(b)
}
