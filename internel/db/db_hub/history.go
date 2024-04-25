package db_hub

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
	ID         uint   `gorm:"column:id;primary_key;AUTO_INCREMENT" json:"id"`
	DeviceType string `gorm:"column:device_type" json:"device_type"`
	DeviceName string `gorm:"column:device_name" json:"device_name"`
	DeviceId   string `gorm:"column:device_id" json:"device_id"`
	UserId     string `gorm:"column:user_id;NOT NULL" json:"user_id"`
	Action     string `gorm:"column:action" json:"action"`
	Timestamp  int    `gorm:"column:timestamp" json:"timestamp"`
	Data       string `gorm:"column:data" json:"data"`
	CreatedAt  string `gorm:"column:created_at;<-:false;default:null" json:"created_at"` // 数据入库时间
	UpdatedAt  string `gorm:"column:updated_at;<-:false;default:null" json:"updated_at"` // 数据修改时间
}

func (m *Device) TableName() string { return "device" }
