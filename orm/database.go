package orm

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	GormDB *gorm.DB
)

func InitDatabase() {
	var err error
	GormDB, err = gorm.Open(sqlite.Open("podcastify.db"), &gorm.Config{
		PrepareStmt: true,
	})
	if err != nil {
		panic("failed to connect database " + err.Error())
	}
	err = GormDB.AutoMigrate(
		&Podcast{},
	)
	if err != nil {
		panic("failed to migrate database " + err.Error())
	}
}
