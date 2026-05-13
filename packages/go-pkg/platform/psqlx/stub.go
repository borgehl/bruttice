package psqlx

import (
	"github.com/borgehl/bruttice/packages/go-pkg/platform/util"
	"fmt"
	"os"
	"regexp"
	"strconv"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

func NewDBFromEnvVars() (*gorm.DB, error) {
	config := NewConfig()

	dbLogLevelMap := map[string]logger.LogLevel{
		"debug":  logger.Info,
		"info":   logger.Info,
		"warn":   logger.Warn,
		"error":  logger.Error,
		"silent": logger.Silent,
	}
	logMode := logger.Default.LogMode(dbLogLevelMap[util.GetLogLevel("DB_LOG_LEVEL")])

	db, err := gorm.Open(postgres.Open(config.GetConnectionString()), &gorm.Config{
		Logger:                 logMode,
		FullSaveAssociations:   false,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, err
	}

	dbRef, err := db.DB()
	if err != nil {
		return nil, err
	}
	if err = dbRef.Ping(); err != nil {
		return nil, err
	}

	dbRef.SetMaxIdleConns(config.MaxIdleConnections)
	dbRef.SetMaxOpenConns(config.MaxOpenConnections)

	if config.SaveSQLAfterExecution {
		cb := dbStatementCallback(config)
		db.Callback().Query().Register("*", cb)
		db.Callback().Create().Register("*", cb)
		db.Callback().Update().Register("*", cb)
		db.Callback().Delete().Register("*", cb)
		db.Callback().Row().Register("*", cb)
		db.Callback().Raw().Register("*", cb)
	}

	return db, nil
}

func dbStatementCallback(config *Config) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		stmt := db.Statement
		sqlQuery := db.Dialector.Explain(stmt.SQL.String(), stmt.Vars...)
		db.Set(config.CallbackSqlKey, sqlQuery)
	}
}
