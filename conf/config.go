package conf

import (
	"gopkg.in/ini.v1"
	db "recache/internal/middleware/mysql/db"
	"strings"
)

var (
	AppMode  string
	HttpPort string

	DB         string
	DbHost     string
	DbPort     string
	DbUser     string
	DbPassword string
	DbName     string

	Policy string
)

func Init() {
	// 本地读取环境变量
	file, err := ini.Load("C:/Users/Administrator/GolandProjects/ReniaoCache/conf/config.ini")
	if err != nil {
		panic(err)
	}
	LoadServer(file)
	LoadDB(file)
	dsn := strings.Join([]string{DbUser, ":", DbPassword, "@tcp(", DbHost, ":", DbPort, ")/", DbName, "?charset=utf8mb4&parseTime=true&loc=Local"}, "")
	db.Database(dsn)
}

func LoadDB(file *ini.File) {
	DB = file.Section("db").Key("DB").String()
	DbHost = file.Section("db").Key("DbHost").String()
	DbPort = file.Section("db").Key("DbPort").String()
	DbUser = file.Section("db").Key("DbUser").String()
	DbPassword = file.Section("db").Key("DbPassword").String()
	DbName = file.Section("db").Key("DbName").String()
}

func LoadServer(file *ini.File) {
	Policy = file.Section("service").Key("Policy").String()
	AppMode = file.Section("service").Key("AppMode").String()
	HttpPort = file.Section("service").Key("HttpPort").String()
}
