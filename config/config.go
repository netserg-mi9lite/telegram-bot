package config

import (
	"net"
	"os"

	"github.com/joho/godotenv"
)

const (
	AppName    = "Gateway Monitor"
	AppVersion = "1.0.0"
)

const SuperAdminUID int64 = 533098160

type Config struct {
	Token  string
	DBPath string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	return &Config{
		Token:  os.Getenv("TOKEN"),
		DBPath: os.Getenv("DB_PATH"),
	}, nil
}

func (c *Config) IsAdmin(userID int64) bool {
	return userID == SuperAdminUID
}

func GetServerIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "unknown"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "unknown"
}
