package config

import (
	"os"

	"github.com/joho/godotenv"
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
