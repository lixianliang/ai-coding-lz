package dbutil

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

type Config struct {
	Host           string `json:"host"`
	Port           int    `json:"port"`
	User           string `json:"user"`
	Password       string `json:"password"`
	Database       string `json:"database"`
	MaxIdleConns   int    `json:"max_idle_conns"`
	MaxIdleTimeSec int    `json:"max_idle_time_sec"`
	EnableLog      bool   `json:"enable_log"`
}

// NewDatabase 初始化数据库
func NewDatabase(conf Config) (*gorm.DB, error) {
	if conf.Host == "" || conf.User == "" || conf.Database == "" {
		return nil, errors.New("invalid host or user or database")
	}

	if conf.Port == 0 {
		conf.Port = 3306
	}

	if err := ensureDatabaseExists(conf); err != nil {
		return nil, fmt.Errorf("failed to ensure database exists: %w", err)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		conf.User,
		conf.Password,
		conf.Host,
		conf.Port,
		conf.Database,
	)

	gormConfig := &gorm.Config{
		// 默认不打印 gorm 日志
		Logger: glogger.Default.LogMode(glogger.Silent),
	}
	if conf.EnableLog {
		gormConfig.Logger = glogger.Default.LogMode(glogger.Info)
	}
	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(conf.MaxIdleConns)
	sqlDB.SetConnMaxIdleTime(time.Duration(conf.MaxIdleTimeSec) * time.Second)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return db, nil
}

func ensureDatabaseExists(conf Config) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8mb4&parseTime=True&loc=Local",
		conf.User,
		conf.Password,
		conf.Host,
		conf.Port,
	)

	gormConfig := &gorm.Config{
		Logger: glogger.Default.LogMode(glogger.Silent),
	}
	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL server: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}
	defer sqlDB.Close()

	createDBSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", conf.Database)
	if err := db.Exec(createDBSQL).Error; err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	return nil
}
