package mysql

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

type Student struct {
	gorm.Model
	Name  string `json:"name"`
	Score string `json:"score"`
}

func Database(dsn string) {
	var ormLogger logger.Interface
	ormLogger = logger.Default.LogMode(logger.Info)
	_db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: ormLogger,
	})
	if err != nil {
		panic(err.Error())
	}
	err = _db.AutoMigrate(&Student{})
	if err != nil {
		panic(err.Error())
	}
	DB = _db
}
