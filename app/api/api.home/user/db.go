package user

var TableUser = "user"

type DBUser struct {
	ID         int    `gorm:"column:id;primary_key;AUTO_INCREMENT" json:"id"`
	Phone      string `gorm:"column:phone" json:"phone"`
	HaAddress1 string `gorm:"column:ha_address_1;NOT NULL" json:"ha_address_1"`
	HaAddress2 string `gorm:"column:ha_address_2;NOT NULL" json:"ha_address_2"`
	CreatedAt  string `gorm:"column:created_at;<-:false;default:null" json:"created_at"` // 数据入库时间
	UpdatedAt  string `gorm:"column:updated_at;<-:false;default:null" json:"updated_at"` // 数据修改时间
}

func (m *DBUser) TableName() string {
	return "user"
}
