package langchain

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"vinesai/internel/ava"
	"vinesai/internel/db"
	"vinesai/internel/lib/tuyago"
	"vinesai/internel/x"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// 同步设备
// 检查信号强度
type syncDevices struct{ CallbacksHandler LogHandler }

func (s *syncDevices) Name() string {
	return "sync_devices"
}

func (s *syncDevices) Description() string {
	return `用于同步设备,初始化分组，分组创建`
}

// 设备数量通知，离线设备通知，场景数量通知，
func (s *syncDevices) Call(ctx context.Context, input string) (string, error) {
	var c = fromCtx(ctx)
	_, _, err := syncDevicesForSummary(c, getHomeId(c))
	if err != nil {
		c.Error(err)
		return "", err
	}

	return "设备同步成功", nil
}

// 同步用户设备信息
// 设备名称数组
// 设备名：详细设备数据
// 只同步可控设备

func syncDevicesForSummary(c *ava.Context, homeId string) ([]string, map[string]*mgoDocDevice, error) {

	//todo 只允许owner_id所有者执行

	//分布式锁
	lock := db.RedisLock.NewMutex(redisKeyTuyaLockSyncDevices + homeId)
	err := lock.Lock()
	if err != nil {
		c.Error(err)
		return nil, nil, err
	}

	defer lock.Unlock()

	//获取房间信息
	var roomResp = &roomInfo{}

	err = tuyago.Get(c, fmt.Sprintf("/v1.0/homes/%s/rooms", homeId), roomResp)

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
	var devicesNameMap = make(map[string]*mgoDocDevice, 20)

	var deviceDoc = make([]interface{}, 0, 30)
	var devicesMap = make(map[string]*mgoDocDevice, 30)

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
			ava.Debugf("get mgoDocDevice list from room fail |data=%v |id=%v", tmpDevicesResp, tmpRoom.RoomID)
			continue
		}

		c.Debugf("所有设备 ｜homeId=%s |rooName=%s |tmpResp=%v", homeId, tmpRoom.Name, x.MustMarshal2String(tmpDevicesResp))

		if len(tmpDevicesResp.Result) == 0 {
			continue
		}

		for ii := range tmpDevicesResp.Result {
			tmpDeviceData := tmpDevicesResp.Result[ii]

			//判断设备名称中是否包含房间位置
			if !strings.Contains(tmpDeviceData.Name, tmpRoom.Name) {
				tmpDeviceData.Name = tmpRoom.Name + tmpDeviceData.Name
			}

			tmpDeviceData.RoomName = tmpRoom.Name
			tmpDeviceData.RoomId = tmpRoom.RoomID
			tmpDeviceData.HomeId = homeId
			tmpDeviceData.CategoryName = getCategoryName(tmpDeviceData.Category)
			deviceDoc = append(deviceDoc, tmpDeviceData)
			devicesMap[tmpDeviceData.ID] = tmpDeviceData

			//如果设备品类不在控制范围内，则不添
			if getCategoryName(tmpDeviceData.Category) == "" {
				continue
			}

			//for summary tool
			devicesName = append(devicesName, tmpDeviceData.Name)
			devicesNameMap[tmpDeviceData.Name] = tmpDeviceData
		}
	}

	//删除所有数据，重新插入
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collectionDevices := db.Mgo.Collection(mgoCollectionNameDevice)
	result, err := collectionDevices.DeleteMany(ctx, bson.M{"homeid": homeId})
	if err != nil {
		c.Error(err)
		return nil, nil, err
	}

	c.Debugf("sync devices delete result %v", result)

	_, err = collectionDevices.InsertMany(ctx, deviceDoc)
	if err != nil {
		c.Error(err)
		return nil, nil, err
	}

	//家庭设备支持的场景
	//删除所有数据，重新插入
	collectionCodes := db.Mgo.Collection(mgoCollectionNameCodes)
	resultCodes, err := collectionCodes.DeleteMany(ctx, bson.M{"homeid": homeId})
	if err != nil {
		c.Error(err)
		return nil, nil, err
	}

	c.Debugf("resultCodes delete result %v", resultCodes)

	//所有家庭场景动作指令
	var codes homeCodeAndStatusResp
	err = tuyago.Get(c, fmt.Sprintf("/v1.0/homes/%s/enable-linkage/codes", homeId), &codes)
	if err != nil {
		c.Error(err)
		return nil, nil, err
	}

	if len(codes.Result) == 0 || !codes.Success {
		return nil, nil, fmt.Errorf("no devices |homeid=%s", homeId)
	}

	var codesInsertDoc []homeFunctionAndStatus
	for i := range codes.Result {
		tmp := codes.Result[i]
		if v, ok := devicesMap[tmp.DeviceId]; ok {
			if v.RoomName == "" {
				continue
			}
			tmp.Name = v.Name
			tmp.HomeId = v.HomeId
			tmp.RoomName = v.RoomName
			tmp.ProductID = v.ProductID
			tmp.Category = v.Category
			codesInsertDoc = append(codesInsertDoc, tmp)
		}
	}

	_, err = collectionCodes.InsertMany(ctx, codesInsertDoc)
	if err != nil {
		c.Error(err)
		return nil, nil, err
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

	//err = syncGroup(c, homeId, devicesNameMap)
	//if err != nil {
	//	c.Error(err)
	//	return nil, nil, err
	//}

	return devicesName, devicesNameMap, nil
}

func removeWhitespace(input string) string {
	// 去除空格、换行符、回车符、制表符等空白字符
	result := strings.ReplaceAll(input, " ", "")
	result = strings.ReplaceAll(result, "\n", "")
	result = strings.ReplaceAll(result, "\r", "")
	result = strings.ReplaceAll(result, "\t", "")
	return result
}

//type devicesGroup struct {
//	SpaceId   string `json:"space_id"`
//	Name      string `json:"name"`
//	DeviceIds string `json:"device_ids"`
//	ProductId string `json:"product_id"`
//}
//
//func getGroupName(roomName, productName string) string {
//	return roomName + productName
//}
//
//var lightCatetoryMap = map[string]string{
//	"dj":  "灯具，包括灯带，筒灯，射灯，轨道灯,双色灯，多色灯等",
//	"xdd": "吸顶灯",
//	"fwd": "氛围灯",
//	"dc":  "灯串",
//	"dd":  "灯带",
//}
//
//var switchCatetoryMap = map[string]string{
//	"kg":   "开关",
//	"cz":   "插座",
//	"pc":   "插排",
//	"cjkg": "场景开关",
//	"tgkg": "调光开关",
//}

// 同步分组,涂鸦目前接口智能分组wifi设备
// 只针对，开关，插排，灯光分组
// 灯光针对productId相同的为一组
// 插排，开关按照category分为一组
//func syncGroup(c *ava.Context, homeId string, devices map[string]*mgoDocDevice) error {
//
//	var group = make(map[string]*devicesGroup)
//	//判断设备群组是否已经存在
//	for _, d := range devices {
//		if _, exist := group[d.ProductId]; exist {
//			group[d.ProductId].DeviceIds = group[d.ProductId].DeviceIds + "," + d.Id
//			continue
//		}
//
//		//灯光分组
//		if _, ok := lightCatetoryMap[d.Category]; ok {
//			var nameType = "单色灯"
//			for i := range d.Status {
//				if strings.Contains(d.Status[i].Code, "music") {
//					nameType = "音乐灯"
//					break
//				}
//
//				if strings.Contains(d.Status[i].Code, "colour") {
//					nameType = "彩灯"
//					break
//				}
//
//				if strings.Contains(d.Status[i].Code, "temp") {
//					nameType = "三色灯"
//					break
//				}
//			}
//
//			groupName := getGroupName(d.roomName, nameType+"群组")
//			group[d.ProductId] = &devicesGroup{
//				SpaceId:   homeId,
//				Name:      groupName,
//				DeviceIds: d.Id,
//				ProductId: d.ProductId,
//			}
//
//			continue
//		}
//
//		//插排，开关分组
//		if v, ok := switchCatetoryMap[d.Category]; ok {
//			var nameType = v
//			var count = 0
//			for i := range d.Status {
//				if strings.Contains(d.Status[i].Code, "switch_mode") {
//					count++
//				}
//			}
//
//			if count > 0 {
//				nameType = strconv.Itoa(count) + "开" + v
//			}
//
//			groupName := getGroupName(d.roomName, nameType+"群组")
//			group[d.ProductId] = &devicesGroup{
//				SpaceId:   homeId,
//				Name:      groupName,
//				DeviceIds: d.Id,
//				ProductId: d.ProductId,
//			}
//		}
//
//	}
//
//	return nil
//}
