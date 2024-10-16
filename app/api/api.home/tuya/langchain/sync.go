package langchain

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"vinesai/internel/ava"
	"vinesai/internel/db"
	"vinesai/internel/lib/tuyago"
	"vinesai/internel/x"
)

// 同步设备
// 检查信号强度
type syncDevices struct{ CallbacksHandler LogHandler }

func (s *syncDevices) Name() string {
	return "sync_devices"
}

func (s *syncDevices) Description() string {
	return `用于同步设备`
}

// 设备数量通知，离线设备通知，场景数量通知，
func (s *syncDevices) Call(ctx context.Context, input string) (string, error) {
	var c = fromCtx(ctx)
	_, _, err := syncDevicesForSummary(c, getHomeId(c))
	if err != nil {
		c.Error(err)
		return "", err
	}

	setSummaryMsg(c, "设备同步成功")

	return "设备同步成功", doneExitError
}

// 同步用户设备信息
// 设备名称数组
// 设备名：详细设备数据
func syncDevicesForSummary(c *ava.Context, homeId string) ([]string, map[string]*device, error) {

	//获取房间信息
	var roomResp = &roomInfo{}

	err := tuyago.Get(c, fmt.Sprintf("/v1.0/homes/%s/rooms", homeId), roomResp)

	if err != nil {
		c.Error(err)
		return nil, nil, err
	}

	if !roomResp.Success {
		return nil, nil, errors.New("获取房间信息失败")
	}

	c.Debugf("roomInfo: %v", roomResp)

	if len(roomResp.Result.Rooms) == 0 {
		return nil, nil, errors.New("请创建房间，并将设备添加到房间中")
	}

	var devicesName = make([]string, 0, 20)
	var devicesNameMap = make(map[string]*device, 20)

	//遍历房间获取设备
	for i := range roomResp.Result.Rooms {
		var tmpRoom = roomResp.Result.Rooms[i]

		var tmpDevicesResp = &deviceListResp{}

		err = tuyago.Get(c, fmt.Sprintf("/v1.0/homes/%s/rooms/%d/devices", homeId, tmpRoom.RoomID), tmpDevicesResp)

		if err != nil {
			ava.Error(err)
			return nil, nil, err
		}

		if !tmpDevicesResp.Success {
			ava.Debugf("get device list from room fail |data=%v |id=%v", tmpDevicesResp, tmpRoom.RoomID)
			continue
		}

		c.Debugf("所有设备 ｜homeId=%s |rooName=%s |tmpResp=%v", homeId, tmpRoom.Name, x.MustMarshal2String(tmpDevicesResp))

		if len(tmpDevicesResp.Result) == 0 {
			continue
		}

		for ii := range tmpDevicesResp.Result {
			tmpDeviceData := tmpDevicesResp.Result[ii]

			//如果设备品类不在控制范围内，则不添
			if getCategoryName(tmpDeviceData.Category) == "" {
				continue
			}

			//判断设备名称中是否包含房间位置
			if !strings.Contains(tmpDeviceData.Name, tmpRoom.Name) {
				//修改设备名称
				//改为修改自己缓存设备的名称
				//涂鸦api设备数量有限制，不修改app上的涂鸦设备名称
				//renameBody := &struct {
				//	Name string `json:"name"`
				//}{
				//	Name: tmpRoom.Name + tmpDeviceData.Name,
				//}
				//renameResp := &struct {
				//	Result bool `json:"result"`
				//}{}
				//err = tuyago.Put(c, "/v1.0/devices/"+tmpDeviceData.Id, renameBody, renameResp)
				//if err != nil {
				//	c.Error(err)
				//	break
				//}
				//
				//if !renameResp.Result {
				//	break
				//}

				tmpDeviceData.Name = tmpRoom.Name + tmpDeviceData.Name

			}

			tmpDeviceData.roomName = tmpRoom.Name
			devicesName = append(devicesName, tmpDeviceData.Name)
			devicesNameMap[tmpDeviceData.Name] = tmpDeviceData
		}
	}

	err = db.GRedis.Set(
		context.Background(),
		redisKeyTuyaSummaryDeviceName+homeId,
		x.MustMarshal2String(devicesName),
		0).Err()
	if err != nil {
		c.Error(err)
		return nil, nil, err
	}

	err = db.GRedis.Set(
		context.Background(),
		redisKeyTuyaSummaryDeviceNameMap+homeId,
		x.MustMarshal2String(devicesNameMap),
		0).Err()
	if err != nil {
		c.Error(err)
		return nil, nil, err
	}

	err = syncGroup(c, homeId, devicesNameMap)
	if err != nil {
		c.Error(err)
		return nil, nil, err
	}

	return devicesName, devicesNameMap, nil
}

type devicesGroup struct {
	SpaceId   string `json:"space_id"`
	Name      string `json:"name"`
	DeviceIds string `json:"device_ids"`
	ProductId string `json:"product_id"`
}

type devicesGroupResp struct {
	Result struct {
		CreateTime       int64  `json:"create_time"`
		GroupID          int    `json:"group_id"`
		GroupName        string `json:"group_name"`
		IconURL          string `json:"icon_url"`
		LocalID          string `json:"local_id"`
		SpaceID          string `json:"space_id"`
		TaskExpireSecond int    `json:"task_expire_second"`
		TaskID           string `json:"task_id"`
		TaskType         int    `json:"task_type"`
		UpdateTime       int64  `json:"update_time"`
	} `json:"result"`
	Success bool   `json:"success"`
	T       int64  `json:"t"`
	Tid     string `json:"tid"`
}

type allGroupResp struct {
	Result struct {
		Count    int `json:"count"`
		DataList []struct {
			Time    int    `json:"time"`
			Id      int    `json:"id"`
			Name    string `json:"name"`
			IconURL string `json:"icon_url"`
			SpaceID string `json:"space_id"`
		} `json:"data_list"`
		PageNo   int `json:"page_no"`
		PageSize int `json:"page_size"`
	} `json:"result"`
	Success bool   `json:"success"`
	T       int64  `json:"t"`
	Tid     string `json:"tid"`
}

func getGroupName(roomName, productName string) string {
	return roomName + productName
}

var lightCatetoryMap = map[string]string{
	"dj":  "灯具，包括灯带，筒灯，射灯，轨道灯,双色灯，多色灯等",
	"xdd": "吸顶灯",
	"fwd": "氛围灯",
	"dc":  "灯串",
	"dd":  "灯带",
}

var switchCatetoryMap = map[string]string{
	"kg":   "开关",
	"cz":   "插座",
	"pc":   "插排",
	"cjkg": "场景开关",
	"tgkg": "调光开关",
}

// 同步分组
// 只针对，开关，插排，灯光分组
// 灯光针对productId相同的为一组
// 插排，开关按照category分为一组
func syncGroup(c *ava.Context, homeId string, devices map[string]*device) error {

	var group = make(map[string]*devicesGroup)
	//判断设备群组是否已经存在
	for _, d := range devices {
		if _, exist := group[d.ProductId]; exist {
			group[d.ProductId].DeviceIds = group[d.ProductId].DeviceIds + "," + d.Id
			continue
		}

		//灯光分组
		if _, ok := lightCatetoryMap[d.Category]; ok {
			var nameType = "单色灯"
			for i := range d.Status {
				if strings.Contains(d.Status[i].Code, "music") {
					nameType = "音乐灯"
					break
				}

				if strings.Contains(d.Status[i].Code, "colour") {
					nameType = "彩灯"
					break
				}

				if strings.Contains(d.Status[i].Code, "temp") {
					nameType = "三色灯"
					break
				}
			}

			groupName := getGroupName(d.roomName, nameType+"群组")
			group[d.ProductId] = &devicesGroup{
				SpaceId:   homeId,
				Name:      groupName,
				DeviceIds: d.Id,
				ProductId: d.ProductId,
			}

			continue
		}

		//插排，开关分组
		if v, ok := switchCatetoryMap[d.Category]; ok {
			var nameType = v
			var count = 0
			for i := range d.Status {
				if strings.Contains(d.Status[i].Code, "switch_mode") {
					count++
				}
			}

			if count > 0 {
				nameType = strconv.Itoa(count) + "开" + v
			}

			groupName := getGroupName(d.roomName, nameType+"群组")
			group[d.ProductId] = &devicesGroup{
				SpaceId:   homeId,
				Name:      groupName,
				DeviceIds: d.Id,
				ProductId: d.ProductId,
			}
		}

	}

	//str, err := db.GRedis.Get(context.Background(), redisKeyTuyaSyncDevices+homeId).Result()
	//if err != nil && !errors.Is(err, redis.Nil) {
	//	c.Error(err)
	//	return err
	//}

	var allResp = &allGroupResp{}

	//if str != "" {
	//	err = x.MustNativeUnmarshal([]byte(str), allResp)
	//	if err != nil {
	//		c.Error(err)
	//		return nil
	//	}
	//}

	//if len(allResp.Result.DataList) == 0 {
	//获取所有群组信息
	err := tuyago.Get(c, "/v2.0/cloud/thing/group?page_no=1&page_size=50&space_id="+homeId, allResp)
	if err != nil {
		c.Error(err)
		return err
	}

	//err = db.GRedis.Set(context.Background(), redisKeyTuyaSyncDevices+homeId, x.MustMarshal2String(allResp), 0).Err()
	//if err != nil {
	//	c.Error(err)
	//	return err
	//}
	//}

	//创建分组
	for _, d := range group {
		var isExist = false
		for i := range allResp.Result.DataList {
			g := allResp.Result.DataList[i]
			if g.Name == d.Name {
				isExist = true
			}
		}

		if isExist {
			continue
		}

		var resp = &devicesGroupResp{}
		err = tuyago.Post(c, "/v2.0/cloud/thing/group", d, resp)
		if err != nil {
			c.Error(err)
			continue
		}

		//todo 防止并发冲突，后期优化为消息队列
	}

	return nil
}
