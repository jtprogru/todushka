package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server *ServerConfig
	DB     *DBConfig
}

type ServerConfig struct {
	Addr     string
	LogLevel string
}

type DBConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
}

func New() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.todushka")
	if err := viper.ReadInConfig(); err != nil {
		panic("can't load config")
	}

	sc := viper.Sub("server")
	if sc == nil {
		panic("server configuration not found")
	}

	dbc := viper.Sub("db")
	if sc == nil {
		panic("db configuration not found")
	}

	return &Config{
		Server: newServerConfig(sc),
		DB:     newDBConfig(dbc),
	}
}

func newServerConfig(v *viper.Viper) *ServerConfig {
	return &ServerConfig{
		Addr:     v.GetString("addr"),
		LogLevel: v.GetString("log_level"),
	}
}

func newDBConfig(v *viper.Viper) *DBConfig {
	return &DBConfig{
		Host:     v.GetString("host"),
		Port:     v.GetString("port"),
		Username: v.GetString("username"),
		Password: v.GetString("password"),
		DBName:   v.GetString("db_name"),
		SSLMode:  v.GetString("ssl_mode"),
	}
}
