package controlcookie

import (
	"github.com/borgehl/bruttice/packages/go-pkg/platform/dbx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/logx"
)

type CookieStore struct {
	database       *dbx.DatabaseClient
	logger         *logx.Logger
	SessionCookies map[string]struct{}
	cookieSalt     string
}

func NewCookieStore(database *dbx.DatabaseClient, cookieSalt string) *CookieStore {
	return &CookieStore{
		database:       database,
		logger:         database.Logger,
		SessionCookies: make(map[string]struct{}),
		cookieSalt:     cookieSalt,
	}
}
