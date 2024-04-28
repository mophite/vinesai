package db_hub

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var TableMessageHistory = "message_history"
var TableDeviceList = "device"

var DatabaseMongoVinesai = "vinesai"
var CollectionDevice = "device"

type MessageHistory struct {
	ID         uint   `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	Option     int    `gorm:"column:option"`
	MerchantId string `gorm:"column:merchant_id"`
	Message    string `gorm:"column:message;NOT NULL"`
	Tip        string `gorm:"column:tip"`
	Exp        string `gorm:"column:exp"`
	Resp       string `gorm:"column:resp"`
	Identity   string `gorm:"column:identity"`
	CreatedAt  string `gorm:"column:created_at;<-:false;default:null"` // 数据入库时间
	UpdatedAt  string `gorm:"column:updated_at;<-:false;default:null"` // 数据修改时间
}

func (m *MessageHistory) TableName() string {
	return "message_history"
}

type Device struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`                //id
	DeviceType int                `bson:"device_type" json:"device_type"`         //设备类型
	DeviceZn   string             `bson:"device_zn" json:"device_zn"`             //设备中文j名称
	DeviceEn   string             `bson:"device_en" json:"device_en"`             //设备英文名称
	DeviceID   string             `bson:"device_id" json:"device_id"`             //设备id
	DeviceDes  string             `bson:"device_des" json:"device_des"`           //设备描述
	Version    string             `bson:"version" json:"version"`                 //设备版本
	UserID     string             `bson:"user_id" json:"user_id"`                 //用户id
	Switch     int                `bson:"switch" json:"switch"`                   //开关，1关，2表示开
	Ip         string             `bson:"ip" json:"ip"`                           //ip
	Wifi       string             `bson:"wifi" json:"wifi"`                       //wifi名称
	CreatedAt  int64              `bson:"created_at,omitempty" json:"created_at"` //创建时间
	UpdatedAt  int64              `bson:"updated_at,omitempty" json:"updated_at"` //更新时间
}

type DeviceAdaptor interface {
	Adaptor2Device() *Device                     //将原来的设备数据转换为通用设备数据
	Adaptor2Native(device *Device) DeviceAdaptor //将通用数据转换为原来的设备数据
}

// 插座二代
// http://help.vi.geek-open.com/work/default/article/37
type SocketMiniV2 struct {
	MessageID string `json:"messageId" bson:"messageId"`
	MAC       string `json:"mac" bson:"mac"`
	Type      string `json:"type" bson:"type"`
	Version   string `json:"version" bson:"version"`
	Key       int    `json:"key" bson:"key"`
	IP        string `json:"ip" bson:"ip"`
	SSID      string `json:"ssid" bson:"ssid"`
}

func (s *SocketMiniV2) Adaptor2Native(device *Device) DeviceAdaptor {
	return &SocketMiniV2{
		MAC:     device.DeviceID,
		Type:    device.DeviceEn,
		Version: device.Version,
		Key:     device.Switch - 1,
	}
}

func (s *SocketMiniV2) Adaptor2Device() *Device {
	return &Device{
		DeviceType: 1,
		DeviceZn:   "智能插座",
		DeviceEn:   s.Type,
		DeviceID:   s.MAC,
		DeviceDes:  "卧室插座", //暂时写死，后期由管理界面输入
		Version:    s.Version,
		Switch:     s.Key + 1,
		Ip:         s.IP,
		Wifi:       s.SSID,
		UserID:     "123",
		CreatedAt:  time.Now().UnixMilli(),
		UpdatedAt:  time.Now().UnixMilli(),
	}
}

func Device2Adaptor(device *Device) (DeviceAdaptor, error) {

	switch device.DeviceType {
	case 1:
		return &SocketMiniV2{
			Type: "event",
			Key:  device.Switch - 1,
		}, nil
	}

	return nil, errors.New("no such device")
}
