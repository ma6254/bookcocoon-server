package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

const (
	SQLITE_DRIVER = "sqlite"
	MYSQL_DRIVER  = "mysql"
)

type Log struct {
	Level string `yaml:"level"`
	Dir   string `yaml:"dir"`
}

type Server struct {
	HttpAddr string `yaml:"http_addr"`
}

type SQLiteDataBase struct {
	File string `yaml:"file"`
}

type MySQLDataBase struct {
	Host     string `yaml:"host"`
	Port     uint32 `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type DataBase struct {
	Salt   string         `yaml:"salt"`
	Driver string         `yaml:"driver"`
	SQLite SQLiteDataBase `yaml:"sqlite"`
	MySQL  MySQLDataBase  `yaml:"mysql"`
}

type Config struct {
	// Server configuration
	Log      Log      `yaml:"log"`
	Server   Server   `yaml:"server"`
	DataBase DataBase `yaml:"database"`
}

// NewConfig 创建默认配置
func NewConfig() *Config {
	return &Config{
		Log: Log{
			Level: "info",
			Dir:   "./",
		},
		Server: Server{
			HttpAddr: "127.0.0.1:7890",
		},
		DataBase: DataBase{
			Driver: "sqlite",
			SQLite: SQLiteDataBase{
				File: "./data.db",
			},
		},
	}
}

// NewConfigPath 从指定路径加载配置文件
func NewConfigPath(file_path string) (*Config, error) {

	data, err := os.ReadFile(file_path)
	if err != nil {
		return nil, err
	}

	c := NewConfig()
	err = yaml.Unmarshal(data, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}
