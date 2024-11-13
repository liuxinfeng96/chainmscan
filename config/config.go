package config

import (
	"chainmscan/db"
	"chainmscan/logger"
	"errors"
	"flag"
	"fmt"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type Config struct {
	ServerPort     string            `mapstructure:"server_port"`
	LogConfig      *logger.LogConfig `mapstructure:"log_config"`
	MysqlConfig    *db.MysqlConfig   `mapstructure:"mysql"`
	GormConfig     *db.GormConfig    `mapstructure:"gorm_config"`
	UploadFilePath string            `mapstructure:"upload_file_path"`
}

const (
	DefaultServerPort     = "9660"
	DefaultUploadFilePath = "./tmp"
)

var configLastChangeTime time.Time

// GetFlagPath --Specify the path and name of the configuration file (flag)
func GetFlagPath() string {
	var configPath string
	flag.StringVar(&configPath, "config", "./conf/config.yaml", "please input the system config file path")
	flag.Parse()
	return configPath
}

// InitConfig --Set config path and file name
func InitConfig(configPath string) (*Config, error) {
	var err error
	var conf Config
	if len(configPath) == 0 {
		configPath = GetFlagPath()
	}

	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile(configPath)
	err = v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	//var conf Config
	err = v.Unmarshal(&conf)
	if err != nil {
		return nil, err
	}

	v.WatchConfig()

	v.OnConfigChange(func(changeEvent fsnotify.Event) {
		if time.Since(configLastChangeTime).Seconds() >= 1 {
			if changeEvent.Op.String() == "WRITE" {
				configLastChangeTime = time.Now()
				err := viper.Unmarshal(&conf)
				if err != nil {
					fmt.Printf("the config hot update failed: [%s]\n", err.Error())
				}
			}
		}
	})

	if conf.LogConfig == nil {
		conf.LogConfig = new(logger.LogConfig)
	}

	if conf.MysqlConfig == nil {
		return nil, errors.New("not found the mysql config")
	}

	if conf.GormConfig == nil {
		return nil, errors.New("not found the gorm config")
	}

	if len(conf.ServerPort) == 0 {
		conf.ServerPort = DefaultServerPort
	}

	if len(conf.UploadFilePath) == 0 {
		conf.UploadFilePath = DefaultUploadFilePath
	}

	return &conf, nil
}
