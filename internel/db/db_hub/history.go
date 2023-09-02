package db_hub

var MessageHistoryTable = "message_history"

type MessageHistory struct {
	ID        uint   `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	Message   string `gorm:"column:message;NOT NULL"`
	Tip       string `gorm:"column:tip"`
	Exp       string `gorm:"column:exp"`
	Resp      string `gorm:"column:resp"`
	HomeID    string `gorm:"column:home_id;NOT NULL"`
	CreatedAt string `gorm:"column:created_at;<-:false;default:null"` // 数据入库时间
	UpdatedAt string `gorm:"column:updated_at;<-:false;default:null"` // 数据修改时间
}

func (m *MessageHistory) TableName() string {
	return "message_history"
}
