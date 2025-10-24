package db

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// User 系统用户表结构体
type User struct {
	ID         int64  `gorm:"primaryKey"`
	Username   string `gorm:"uniqueIndex:uk_username"`
	Password   string
	SuperAdmin uint8 `gorm:"comment:'超级管理员   0：否   1：是'"`
	Status     int8  `gorm:"comment:'状态  0：停用   1：正常'"`
	CreateDate time.Time
	Updater    int64
	Creator    int64
	UpdateDate time.Time
}

// TableName 指定表名
func (User) TableName() string {
	return "sys_user"
}

type UserToken struct {
	ID         int64  `gorm:"primaryKey"`
	UserID     int64  `gorm:"uniqueIndex"`
	Token      string `gorm:"uniqueIndex"`
	ExpireDate time.Time
	UpdateDate time.Time
	CreateDate time.Time
}

// TableName 映射表名
func (UserToken) TableName() string {
	return "sys_user_token"
}

func (db *Database) UserToken(ctx context.Context, token string) (UserToken, error) {
	return gorm.G[UserToken](db.db).Where("token = ?", token).Take(ctx)
}

func (db *Database) User(ctx context.Context, uid int64) (User, error) {
	return gorm.G[User](db.db).Where("id = ?", uid).Take(ctx)
}

func (db *Database) GetAdminID(ctx context.Context) (int64, error) {
	admin, err := gorm.G[User](db.db).Where("super_admin = ?", 1).Take(ctx)
	if err != nil {
		return 0, err
	}
	return admin.ID, nil
}
