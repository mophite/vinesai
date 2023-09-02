package db

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

var GMysql *gorm.DB

func ChaosMysql(dsn string) error {

	var err error
	GMysql, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	sqlDB, err := GMysql.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxOpenConns(1000)
	sqlDB.SetMaxIdleConns(20)
	sqlDB.SetConnMaxLifetime(time.Minute * 10)

	err = sqlDB.Ping()
	if err != nil {
		return err
	}

	return nil
}
