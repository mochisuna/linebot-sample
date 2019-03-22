package config

import (
	"github.com/BurntSushi/toml"
)

// Config all settings
type Config struct {
	Server   Server `toml:"server"`
	DBMaster DB     `toml:"dbm"`
	DBSlave  DB     `toml:"dbs"`
	Line     Line   `toml:"line"`
}

// Server port
type Server struct {
	Port string `toml:"port"`
}

// Line
type Line struct {
	ChannelSecret string `toml:"channel_secret"`
	ChannelToken  string `toml:"channel_token"`
}

// DB database structure
type DB struct {
	Host     string `toml:"host"`
	Port     string `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	DBName   string `toml:"dbname"`
}

// New Config
func New(config *Config, configPath string) error {
	_, err := toml.DecodeFile(configPath, config)
	return err

}
