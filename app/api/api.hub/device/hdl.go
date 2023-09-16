package device

import (
	"net/http"

	"vinesai/internel/ava"
	"vinesai/internel/ipc"
	"vinesai/internel/x"
	"vinesai/proto/phub"
)

type DevicesHub struct{}

func (d *DevicesHub) TransmitControlCommandFile(c *ava.Context, req *phub.ControlPutFileReq, rsp *phub.ControlPutFileRsp) {
	if req.Extra == "" {
		rsp.Code = http.StatusBadRequest
		rsp.Msg = "缺少extra"
		return
	}

	var home struct {
		HomeId string `json:"homeId"`
	}

	err := x.MustUnmarshal(x.StringToBytes(req.Extra), &home)
	if err != nil {
		rsp.Code = http.StatusBadRequest
		rsp.Msg = "参数不正确"
		return
	}

	if home.HomeId == "" {
		rsp.Code = http.StatusBadRequest
		rsp.Msg = "homeId不能为控"
		return
	}

	result, err := asr(req.Body)

	if err != nil {
		c.Error(err)
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = "asr服务错误"
		return
	}

	if result == nil || result.Response == nil || result.Response.Result == nil {
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = "asr服务错误"
		return
	}

	cRsp, err := ipc.Chat2AI(c, &phub.ChatReq{HomeId: home.HomeId, Message: *result.Response.Result})
	if err != nil {
		c.Error(err)
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = x.StatusInternalServerError
		return
	}

	rsp.Code = http.StatusOK
	rsp.Msg = "指令获取成功"

	data := &phub.ControlDevicesData{
		Tip:  cRsp.Data.Tip,
		Exp:  cRsp.Data.Exp,
		Resp: cRsp.Data.Resp,
	}
	rsp.Data = data
}

func (d *DevicesHub) TransmitControlCommand(c *ava.Context, req chan *phub.ControlFileReq, exit chan struct{}) chan *phub.ControlFileRsp {
	var rsp = make(chan *phub.ControlFileRsp)

	go func() {
	QUIT:
		for {
			select {
			case data, ok := <-req:
				if !ok {
					//just break select
					break
				}

				//判断数据是否正确
				if data.FileName == "" || data.FileSize == 0 || len(data.Body) == 0 {
					c.Debug("TransmitControlCommand 没有内容")
					rsp <- &phub.ControlFileRsp{
						Code: http.StatusBadRequest,
						Msg:  "我不太明白你的意思",
					}
					break
				}

				//发送数据进行语音识别,chatgpt处理
				result, err := asr(data.Body)
				if err != nil {
					c.Error(err)
					rsp <- &phub.ControlFileRsp{
						Code: http.StatusBadRequest,
						Msg:  "对不起，我理解不了",
					}
					break
				}

				c.Debug(x.MustMarshal2String(result))

				if result == nil || result.Response == nil || result.Response.Result == nil {
					rsp <- &phub.ControlFileRsp{
						Code: http.StatusBadRequest,
						Msg:  "我出了点问题，等我一下",
					}
					break
				}
				message := *result.Response.Result

				cRsp, err := ipc.Chat2AI(c, &phub.ChatReq{HomeId: data.HomeId, Message: message})
				if err != nil {
					c.Error(err)
					rsp <- &phub.ControlFileRsp{
						Code: http.StatusBadRequest,
						Msg:  "我不明白",
					}
					break
				}

				cd := &phub.ControlDevicesData{
					Tip:  cRsp.Data.Tip,
					Exp:  cRsp.Data.Exp,
					Resp: cRsp.Data.Resp,
				}

				//todo 将处理结果发给三方

				//测试 直接发送人工文字给es
				rsp <- &phub.ControlFileRsp{
					Code: http.StatusOK,
					Msg:  x.StatusOK,
					Data: cd,
				}

			case <-exit:
				//break all
				break QUIT
			}
		}

		close(rsp)
	}()

	return rsp
}

func (d *DevicesHub) ReportDeviceStatus(c *ava.Context, req *phub.DevicesStatusReq, rsp *phub.DevicesStatusRsp) {
	//TODO implement me
	panic("implement me")
}

func (d *DevicesHub) TransmitControlCommandWord(c *ava.Context, req *phub.ControlWordReq, rsp *phub.ControlWordRsp) {
	if req.Message == "" {
		rsp.Code = http.StatusBadRequest
		rsp.Msg = "请输入控制指令"
		return
	}

	if req.HomeId == "" {
		rsp.Code = http.StatusBadRequest
		rsp.Msg = "homeId不能为控"
		return
	}

	cRsp, err := ipc.Chat2AI(c, &phub.ChatReq{HomeId: req.HomeId, Message: req.Message})
	if err != nil {
		c.Error(err)
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = x.StatusInternalServerError
		return
	}

	rsp.Code = http.StatusOK
	rsp.Msg = "指令获取成功"

	data := &phub.ControlDevicesData{
		Tip:  cRsp.Data.Tip,
		Exp:  cRsp.Data.Exp,
		Resp: cRsp.Data.Resp,
	}
	rsp.Data = data
}

func (d *DevicesHub) ExecuteAndReport(c *ava.Context, req *phub.ReportDeviceAttributesReq, rsp *phub.ReportDeviceAttributesRsp) {
	//TODO implement me
	panic("implement me")
}
