package db_hub

import "go.mongodb.org/mongo-driver/bson/primitive"

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
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	DeviceType string             `bson:"device_type" json:"device_type"`
	DeviceName string             `bson:"device_name" json:"device_name"`
	DeviceID   string             `bson:"device_id" json:"device_id"`
	UserID     string             `bson:"user_id" json:"user_id"`
	Action     string             `bson:"action" json:"action"`
	Data       string             `bson:"data" json:"data"`
	Timestamp  int                `bson:"timestamp" json:"timestamp"`
	CreatedAt  string             `bson:"created_at,omitempty" json:"created_at"`
	UpdatedAt  string             `bson:"updated_at,omitempty" json:"updated_at"`
}
