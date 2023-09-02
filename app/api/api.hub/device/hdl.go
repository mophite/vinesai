package device

import (
	"net/http"
	"vinesai/internel/ava"
	"vinesai/internel/ipc"
	"vinesai/internel/x"
	"vinesai/proto/phub"
)

type DevicesHub struct{}

// 设备发现
func (d *DevicesHub) DiscoverDevices(c *ava.Context, req *phub.DiscoverDevicesReq, rsp *phub.DiscoverDevicesRsp) {
	rsp.Code = http.StatusOK
	rsp.Msg = x.StatusOK
	rsp.Data = &phub.DiscoverDevicesData{Endpoints: []*phub.DevicesEndpoint{
		&phub.DevicesEndpoint{
			EndpointId:       "11",
			FriendlyName:     "test",
			Description:      "描述",
			ManufacturerName: "测试",
		},
	}}
}

// 设备控制
func (d *DevicesHub) ControlDevices(c *ava.Context, req *phub.ControlDevicesReq, rsp *phub.ControlDevicesRsp) {
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

	cRsp, err := ipc.Chat2AI(c, &phub.ChatReq{HomeId: req.HomeId, Message: req.Message, DevicesIds: req.DevicesIds})
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

// 设备状态
func (d *DevicesHub) DevicesStatus(c *ava.Context, req *phub.DevicesStatusReq, rsp *phub.DevicesStatusRsp) {
	rsp.Code = http.StatusOK
	rsp.Msg = x.StatusOK
	rsp.Devices = []*phub.DevicesStatusData{
		&phub.DevicesStatusData{
			DeviceId: "123",
			DeviceAttributes: []*phub.DeviceAttributes{
				&phub.DeviceAttributes{
					Name:  "测试",
					Value: "测试",
				}},
		},
	}
}

// 设备上报
func (d *DevicesHub) ReportDeviceAttributes(c *ava.Context, req *phub.ReportDeviceAttributesReq, rsp *phub.ReportDeviceAttributesRsp) {
	c.Infof("ReportDeviceAttributes |req=%v", x.MustMarshal2String(req))

	rsp.Code = http.StatusOK
	rsp.Msg = x.StatusOK
	rsp.Data = []*phub.ReportDeviceAttributesData{
		&phub.ReportDeviceAttributesData{
			DeviceId: "123",
			Status:   "off",
			Message:  "设备关闭中",
		},
	}
}
