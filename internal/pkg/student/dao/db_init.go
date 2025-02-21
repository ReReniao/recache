package dao

import (
	"context"
	"fmt"
	"recache/config"

	"strings"

	logger2 "recache/utils/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var _db *gorm.DB

func InitDB() {
	mConfig := config.Conf.Mysql
	host := mConfig.Host
	port := mConfig.Port
	database := mConfig.Database
	username := mConfig.UserName
	password := mConfig.Password
	charset := mConfig.Charset

	// username:password@tcp(host:port)/database?charset=xx&parseTime=xx&loc=xx
	dsn := strings.Join([]string{username, ":", password, "@tcp(", host, ":", port, ")/", database, "?charset=", charset, "&parseTime=", "true", "&loc=", "Local"}, "")
	err := Database(dsn)
	if err != nil {
		fmt.Println(err)
		logger2.LogrusObj.Error(err)
	}
}

func Database(connStr string) error {
	// var ormLogger logger.Interface
	// if gin.Mode() == "debug" {
	// 	ormLogger = logger.Default.LogMode(logger.Info)
	// } else {
	// 	ormLogger = logger.Default
	// }

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       connStr,
		DefaultStringSize:         256,   // Default length of String type fields
		DisableDatetimePrecision:  true,  // Disable datetime precision
		DontSupportRenameIndex:    true,  // When renaming an index, delete and create a new one
		DontSupportRenameColumn:   true,  // Rename the column with `change`
		SkipInitializeWithVersion: false, // Automatically configure based on version
	}), &gorm.Config{
		// Logger: ormLogger,
		Logger: logger.Default.LogMode(logger.Silent),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})

	if err != nil {
		panic(err)
	}

	// sqlDB, _ := db.DB()
	// sqlDB.SetMaxIdleConns(20)
	// sqlDB.SetMaxOpenConns(100)
	// sqlDB.SetConnMaxLifetime(30 * time.Second)
	_db = db
	migration()
	return err
}

func NewDBClient(ctx context.Context) *gorm.DB {
	db := _db
	return db.WithContext(ctx)
}
