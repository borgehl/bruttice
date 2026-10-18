package psqlx

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	_ "github.com/lib/pq"
)

var dbUrlRegex = regexp.MustCompile("//(.+):(.+)@(.+):(.+)/(.+)")

type Config struct {
	Host                  string
	Port                  int
	Username              string
	Password              string
	Database              string
	SSLMode               string
	MaxIdleConnections    int
	MaxOpenConnections    int
	SaveSQLAfterExecution bool
	CallbackSqlKey        string
}

func NewConfig() *Config {
	port, _ := strconv.Atoi(os.Getenv("POSTGRES_PORT"))
	return &Config{
		Host:     os.Getenv("POSTGRES_HOST"),
		Database: os.Getenv("POSTGRES_DB"),
		Username: os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		Port:     port,
	}
}

func (c *Config) GetConnectionString() string {
	psqlInfo := fmt.Sprintf("host=%s dbname=%s", c.Host, c.Database)
	if c.Port > 0 {
		psqlInfo += fmt.Sprintf(" port=%d", c.Port)
	}
	if c.Username != "" {
		psqlInfo += fmt.Sprintf(" user=%s", c.Username)
	}
	if c.Password != "" {
		psqlInfo += fmt.Sprintf(" password=%s", c.Password)
	}
	if c.SSLMode == "" {
		c.SSLMode = "disable"
	}
	return psqlInfo
}
