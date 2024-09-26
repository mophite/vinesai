package tuya

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"
	"vinesai/internel/x"

	"github.com/tuya/tuya-connector-go/connector"
)

type GetDeviceResponse struct {
	Code    int         `json:"code"`
	Msg     string      `json:"msg"`
	Success bool        `json:"success"`
	Result  interface{} `json:"result"`
	T       int64       `json:"t"`
}

func TestInit(t *testing.T) {
	resp := &GetDeviceResponse{}

	err := connector.MakeGetRequest(
		context.Background(),
		connector.WithAPIUri(fmt.Sprintf("/v1.0/devices/%s", "6c54d390a7e4b4db9cidmn")),
		connector.WithResp(resp))
	if err != nil {
		fmt.Println("err:", err.Error())
		return
	}

	b, _ := json.Marshal(resp)
	fmt.Println(string(b))
}

func TestAfter(t *testing.T) {
	x.TimingwheelAfter(time.Second*5, func() {
		fmt.Println("0000000")
	})

	s := time.After(time.Second * 5)
	select {
	case <-s:
	}
	fmt.Println("----2---")
}
