package db_hub

import (
	"errors"
)

var TableMessageHistory = "message_history"
var TableDeviceList = "device"

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
	ID         uint   `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	DeviceType int    `gorm:"column:device_type" json:"device_type"`                     //设备类型
	DeviceZn   string `gorm:"column:device_zn" json:"device_zn"`                         //设备中文j名称
	DeviceEn   string `gorm:"column:device_en" json:"device_en"`                         //设备英文名称
	DeviceID   string `gorm:"column:device_id" json:"device_id"`                         //设备id
	DeviceDes  string `gorm:"column:device_des" json:"device_des"`                       //设备描述
	Version    string `gorm:"column:version" json:"version"`                             //设备版本
	UserID     string `gorm:"column:user_id" json:"user_id"`                             //用户id
	Control    int    `gorm:"column:control" json:"control"`                             //开关，1关，2表示开
	Ip         string `gorm:"column:ip" json:"ip"`                                       //ip
	Wifi       string `gorm:"column:wifi" json:"wifi"`                                   //wifi名称
	CreatedAt  int64  `gorm:"column:created_at;<-:false;default:null" json:"created_at"` //创建时间
	UpdatedAt  int64  `gorm:"column:updated_at;<-:false;default:null" json:"updated_at"` //更新时间
}

func (m *Device) TableName() string {
	return "device"
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
		Key:     device.Control - 1,
	}
}

func (s *SocketMiniV2) Adaptor2Device() *Device {
	return &Device{
		DeviceType: 1,
		DeviceEn:   s.Type,
		DeviceID:   s.MAC,
		Version:    s.Version,
		Control:    s.Key + 1,
		Ip:         s.IP,
		Wifi:       s.SSID,
		UserID:     "123",
	}
}

func Device2Adaptor(device *Device) (DeviceAdaptor, error) {

	switch device.DeviceType {
	case 1:
		return &SocketMiniV2{
			Type: "event",
			Key:  device.Control - 1,
		}, nil
	}

	return nil, errors.New("no such device")
}
