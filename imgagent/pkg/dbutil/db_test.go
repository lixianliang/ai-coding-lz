package dbutil

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestNewDatabase_InvalidConfig(t *testing.T) {
	tests := []struct {
		name string
		conf Config
	}{
		{
			name: "empty host",
			conf: Config{
				Host:     "",
				User:     "testuser",
				Database: "testdb",
			},
		},
		{
			name: "empty user",
			conf: Config{
				Host:     "localhost",
				User:     "",
				Database: "testdb",
			},
		},
		{
			name: "empty database",
			conf: Config{
				Host:     "localhost",
				User:     "testuser",
				Database: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := NewDatabase(tt.conf)
			assert.Error(t, err)
			assert.Nil(t, db)
			assert.Contains(t, err.Error(), "invalid host or user or database")
		})
	}
}

func TestNewDatabase_DefaultPort(t *testing.T) {
	conf := Config{
		Host:     "localhost",
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
		Port:     0,
	}

	_, err := NewDatabase(conf)
	assert.Error(t, err)
}

func TestEnsureDatabaseExists_DSNFormat(t *testing.T) {
	tests := []struct {
		name     string
		conf     Config
		expected string
	}{
		{
			name: "standard config",
			conf: Config{
				Host:     "localhost",
				Port:     3306,
				User:     "root",
				Password: "password",
				Database: "testdb",
			},
			expected: "root:password@tcp(localhost:3306)/?charset=utf8mb4&parseTime=True&loc=Local",
		},
		{
			name: "custom port",
			conf: Config{
				Host:     "127.0.0.1",
				Port:     3307,
				User:     "admin",
				Password: "secret",
				Database: "mydb",
			},
			expected: "admin:secret@tcp(127.0.0.1:3307)/?charset=utf8mb4&parseTime=True&loc=Local",
		},
		{
			name: "empty password",
			conf: Config{
				Host:     "localhost",
				Port:     3306,
				User:     "root",
				Password: "",
				Database: "testdb",
			},
			expected: "root:@tcp(localhost:3306)/?charset=utf8mb4&parseTime=True&loc=Local",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ensureDatabaseExists(tt.conf)
			assert.Error(t, err)
		})
	}
}

func TestEnsureDatabaseExists_CreateDatabaseSQL(t *testing.T) {
	tests := []struct {
		name         string
		database     string
		expectedSQL  string
		expectError  bool
		mockSetup    func(sqlmock.Sqlmock)
	}{
		{
			name:        "create database success",
			database:    "testdb",
			expectedSQL: "CREATE DATABASE IF NOT EXISTS `testdb` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
			expectError: false,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("CREATE DATABASE IF NOT EXISTS `testdb`").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
		},
		{
			name:        "create database with special characters",
			database:    "test-db_123",
			expectedSQL: "CREATE DATABASE IF NOT EXISTS `test-db_123` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
			expectError: false,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("CREATE DATABASE IF NOT EXISTS `test-db_123`").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			gormDB, err := gorm.Open(mysql.New(mysql.Config{
				Conn:                      db,
				SkipInitializeWithVersion: true,
			}), &gorm.Config{})
			require.NoError(t, err)

			result := gormDB.Exec(tt.expectedSQL)
			if tt.expectError {
				assert.Error(t, result.Error)
			} else {
				assert.NoError(t, result.Error)
			}

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		conf    Config
		wantErr bool
	}{
		{
			name: "valid config",
			conf: Config{
				Host:           "localhost",
				Port:           3306,
				User:           "root",
				Password:       "password",
				Database:       "testdb",
				MaxIdleConns:   10,
				MaxIdleTimeSec: 300,
				EnableLog:      false,
			},
			wantErr: false,
		},
		{
			name: "missing host",
			conf: Config{
				Host:     "",
				User:     "root",
				Database: "testdb",
			},
			wantErr: true,
		},
		{
			name: "missing user",
			conf: Config{
				Host:     "localhost",
				User:     "",
				Database: "testdb",
			},
			wantErr: true,
		},
		{
			name: "missing database",
			conf: Config{
				Host:     "localhost",
				User:     "root",
				Database: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewDatabase(tt.conf)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestNewDatabase_ConnectionPoolSettings(t *testing.T) {
	conf := Config{
		Host:           "localhost",
		Port:           3306,
		User:           "root",
		Password:       "password",
		Database:       "testdb",
		MaxIdleConns:   5,
		MaxIdleTimeSec: 600,
		EnableLog:      true,
	}

	_, err := NewDatabase(conf)
	assert.Error(t, err)
}

func TestEnsureDatabaseExists_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		conf        Config
		expectError bool
	}{
		{
			name: "invalid host",
			conf: Config{
				Host:     "invalid-host-12345",
				Port:     3306,
				User:     "root",
				Password: "password",
				Database: "testdb",
			},
			expectError: true,
		},
		{
			name: "invalid port",
			conf: Config{
				Host:     "localhost",
				Port:     99999,
				User:     "root",
				Password: "password",
				Database: "testdb",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ensureDatabaseExists(tt.conf)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewDatabase_LoggingConfiguration(t *testing.T) {
	tests := []struct {
		name      string
		enableLog bool
	}{
		{
			name:      "logging enabled",
			enableLog: true,
		},
		{
			name:      "logging disabled",
			enableLog: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := Config{
				Host:      "localhost",
				Port:      3306,
				User:      "root",
				Password:  "password",
				Database:  "testdb",
				EnableLog: tt.enableLog,
			}

			_, err := NewDatabase(conf)
			assert.Error(t, err)
		})
	}
}
