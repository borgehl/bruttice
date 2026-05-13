package web

import (
	"github.com/borgehl/bruttice/apps/backend/internal/web/api"
	"github.com/borgehl/bruttice/apps/backend/internal/web/controlcookie"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/dbx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/emailx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/logx"
	"net/http"
)

type Controller struct {
	mux *http.ServeMux

	logger         *logx.Logger
	database       *dbx.DatabaseClient
	email          emailx.EmailClient
	cookieSessions *controlcookie.CookieStore
}

func NewController(logger *logx.Logger, database *dbx.DatabaseClient, mux *http.ServeMux, ingestKey string, cookieSalt string) *Controller {
	ret := &Controller{
		mux:            mux,
		logger:         logger,
		database:       database,
		email:          emailx.NewEmailClient(logger),
		cookieSessions: controlcookie.NewCookieStore(database, cookieSalt),
	}

	apiHandler := api.NewAPIHandler(logger, database, ret.cookieSessions, ingestKey)
	apiHandler.RegisterRoutes(ret.mux)

	return ret
}
