package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server *ServerConfig
	Db     *DbConfig
}

type ServerConfig struct {
	Addr string
}

type DbConfig struct {
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
		Db:     newDbConfig(dbc),
	}
}

func newServerConfig(v *viper.Viper) *ServerConfig {
	return &ServerConfig{
		Addr: v.GetString("addr"),
	}
}

func newDbConfig(v *viper.Viper) *DbConfig {
	return &DbConfig{
		Host:     v.GetString("host"),
		Port:     v.GetString("port"),
		Username: v.GetString("username"),
		Password: v.GetString("password"),
		DBName:   v.GetString("db_name"),
		SSLMode:  v.GetString("ssl_mode"),
	}
}
