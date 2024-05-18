package db

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var GMysql *gorm.DB

func ChaosMysql(dsn string) error {

	var err error
	GMysql, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				//SlowThreshold: time.Second, // Slow SQL threshold
				LogLevel: logger.Info, // Log level
				Colorful: true,        // Disable color
			}),
	})
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
