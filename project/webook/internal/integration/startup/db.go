package startup

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"project/webook/config"
	"project/webook/internal/repository/dao"
)

func InitDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	//db, err := gorm.Open(mysql.Open("root:root@tcp(webook-mysql:3308)/webook"))
	if err != nil {
		panic(err)
	}

	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}

	return db
}
