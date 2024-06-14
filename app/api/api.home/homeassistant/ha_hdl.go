package homeassistant

import (
	"net/http"
	"sync/atomic"
	"time"
	"vinesai/internel/ava"
	"vinesai/internel/x"
	"vinesai/proto/pha"
)

type HomeAssistant struct {
}

func (h *HomeAssistant) Speaker(c *ava.Context, req *pha.SpeakerReq, rsp *pha.SpeakerRsp) {
	//if req.Home == "" {
	//	rsp.Code = http.StatusBadRequest
	//	rsp.Msg = "我不知道你要控制哪个家庭设备"
	//	return
	//}

	if len(req.Messages) == 0 {
		rsp.Code = http.StatusBadRequest
		rsp.Msg = "我听不见"
		return
	}

	//获取音响识别到的最近的一句话
	content := req.Messages[len(req.Messages)-1].Content

	result, err := SelectAI(c, "123", content)
	if err != nil {
		c.Error(err)
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = "奴家有点迷糊"
		return
	}
	id := time.Now().UnixNano()

	for i := range result.Commands {
		command := result.Commands[i]
		//发起设备控制
		callServiceWs(atomic.AddInt64(&id, 1), "123", command)
	}

	rsp.Code = http.StatusOK
	if result.Answer == "" {
		result.Answer = "奴家不太明白"
	}
	rsp.Result = result.Answer
}

// Call todo 用户输入，控制设备
func (h *HomeAssistant) Call(c *ava.Context, req *pha.CallReq, rsp *pha.CallRsp) {
	//if req.Home == "" {
	//	rsp.Code = http.StatusBadRequest
	//	rsp.Msg = "我不知道你要控制哪个家庭设备"
	//	return
	//}

	if req.Message == "" {
		rsp.Code = http.StatusBadRequest
		rsp.Msg = "我听不见"
		return
	}

	result, err := SelectAI(c, "123", req.Message)
	if err != nil {
		c.Error(err)
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = "奴家有点迷糊"
		return
	}

	for i := range result.Commands {
		command := result.Commands[i]
		//发起设备控制
		callServiceHttp("123", command.Service, x.MustMarshal(command.Data))
	}

	rsp.Code = http.StatusOK
	if result.Answer == "" {
		result.Answer = "奴家不太明白"
	}
	rsp.Msg = result.Answer
}

// 获取指令列表
func (h *HomeAssistant) Services(c *ava.Context, req *pha.ServicesReq, rsp *pha.ServicesRsp) {
	//home_id := c.GetHeader("home_id")
	//if home_id == "" || mapHome2Url[home_id] == "" {
	//	rsp.Code = http.StatusBadRequest
	//	rsp.Msg = "header中的home_id不能为空或当前home不存在"
	//	return
	//}
	data, err := getServices(c, "123")
	if err != nil {
		c.Error(err)
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = "没有数据"
		return
	}

	rsp.Data = &pha.ServicesData{
		Services: data,
	}

	rsp.Code = http.StatusOK

}

// 获取设备当前状态
func (h *HomeAssistant) States(c *ava.Context, req *pha.StatesReq, rsp *pha.StatesRsp) {
	//home_id := c.GetHeader("home_id")
	//if home_id == "" || mapHome2Url[home_id] == "" {
	//	rsp.Code = http.StatusBadRequest
	//	rsp.Msg = "header中的home_id不能为空或当前home不存在"
	//	return
	//}

	data, _, err := getStates(c, "123")
	if err != nil {
		c.Error(err)
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = "没有数据"
		return
	}

	rsp.Data = &pha.StatesData{
		States: data,
	}

	rsp.Code = http.StatusOK
}
