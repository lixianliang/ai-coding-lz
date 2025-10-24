package db

import (
	"encoding/hex"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"imgagent/pkg/dbutil"
)

type Database struct {
	db *gorm.DB
}

func NewDatabase(conf dbutil.Config) (*Database, error) {
	db, err := dbutil.NewDatabase(conf)
	if err != nil {
		return nil, err
	}

	database := &Database{
		db: db,
	}

	// 这里可以添加表创建逻辑，需要指定字符集为 utf8mb4，默认为 utf8mb3
	err = db.Set("gorm:table_options", "CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci").AutoMigrate(&Document{}, &Chapter{})

	if err != nil {
		zap.S().Errorf("Failed to auto migrate, err: %v", err)
		database.Close()
		return nil, err
	}

	return database, nil
}

func (db *Database) SetDB(gdb *gorm.DB) {
	db.db = gdb
}

func (db *Database) Close() {
	sqlDB, err := db.db.DB()
	if err != nil {
		zap.S().Errorf("Failed to get db, err: %v", err)
		return
	}
	err = sqlDB.Close()
	if err != nil {
		zap.S().Errorf("Failed to close db, err: %v", err)
	}
}

func MakeUUID() string {
	id := uuid.New()
	return hex.EncodeToString(id[:])
}
