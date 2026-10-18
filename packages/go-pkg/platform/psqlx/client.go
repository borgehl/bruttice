package psqlx

import (
	"github.com/borgehl/bruttice/packages/go-pkg/platform/util"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// New database from environment variables
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
