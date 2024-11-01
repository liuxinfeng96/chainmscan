package db

import (
	"chainmscan/logger"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type MysqlConfig struct {
	User       string `mapstructure:"user"`
	Password   string `mapstructure:"password"`
	Host       string `mapstructure:"host"`
	Port       string `mapstructure:"port"`
	DbName     string `mapstructure:"dbname"`
	Parameters string `mapstructure:"parameters"`
}

func getMysqlDns(mysqlConf *MysqlConfig) string {
	mysqlURL := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s",
		mysqlConf.User, mysqlConf.Password, mysqlConf.Host, mysqlConf.Port, mysqlConf.DbName, mysqlConf.Parameters)
	return mysqlURL
}

func MysqlInit(mysqlConf *MysqlConfig, gormConfig *GormConfig, tableSlice []interface{},
	zaplogger *zap.SugaredLogger) (*gorm.DB, error) {
	var err error

	glogger := logger.NewGormLogger(zaplogger, 200*time.Millisecond, false)
	gormDb, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       getMysqlDns(mysqlConf),
		DefaultStringSize:         256,   // string 类型字段的默认长度
		DisableDatetimePrecision:  true,  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,  // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false, // 根据当前 MySQL 版本自动配置
	}), &gorm.Config{
		Logger: glogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := gormDb.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(gormConfig.MaxIdleConns)
	sqlDB.SetMaxOpenConns(gormConfig.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Second * time.Duration(gormConfig.MaxLifetime))

	if gormConfig.EnableAutoMigrate {
		err = gormDb.AutoMigrate(tableSlice...)
		if err != nil {
			return nil, err
		}
	}

	return gormDb, nil
}
