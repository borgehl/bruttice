package dbx

import (
	"github.com/borgehl/bruttice/packages/go-pkg/platform/logx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/psqlx"

	"gorm.io/gorm"
)

type DatabaseClient struct {
	DB     *gorm.DB
	Logger *logx.Logger
}

func NewDatabaseClient(logger *logx.Logger) *DatabaseClient {
	db, err := psqlx.NewDBFromEnvVars()
	if err != nil {
		panic(err)
	}
	return &DatabaseClient{
		DB:     db,
		Logger: logger,
	}
}
