package schemas

import "gorm.io/gorm"

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&User{}, &Session{}, &Project{}, &TimeEntry{}, &UserSetting{})
}
