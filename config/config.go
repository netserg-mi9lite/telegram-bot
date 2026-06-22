package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Token    string
	AdminIDs []int64
	DBPath   string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	adminIDs := parseAdminIDs(os.Getenv("ADMIN_IDS"))

	return &Config{
		Token:    os.Getenv("TOKEN"),
		AdminIDs: adminIDs,
		DBPath:   os.Getenv("DB_PATH"),
	}, nil
}

func parseAdminIDs(raw string) []int64 {
	var ids []int64
	for _, s := range strings.Split(raw, ",") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		id, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}

func (c *Config) IsAdmin(userID int64) bool {
	for _, id := range c.AdminIDs {
		if id == userID {
			return true
		}
	}
	return false
}
