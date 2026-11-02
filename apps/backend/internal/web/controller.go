package web

import (
	"github.com/borgehl/bruttice/apps/backend/internal/web/api"
	"github.com/borgehl/bruttice/apps/backend/internal/web/controlcookie"
	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/dbx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/emailx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/logx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/pagesx"
	"net/http"
)

type Controller struct {
	mux *http.ServeMux

	logger         *logx.Logger
	database       *dbx.DatabaseClient
	email          *emailx.EmailClient
	cookieSessions *controlcookie.CookieStore
}

func NewController(logger *logx.Logger, database *dbx.DatabaseClient, mux *http.ServeMux, ingestKey string, ingestTrigger api.IngestTrigger, cookieSalt string, site dto.SiteResponse, pages []pagesx.Page, email *emailx.EmailClient) *Controller {
	ret := &Controller{
		mux:            mux,
		logger:         logger,
		database:       database,
		email:          email,
		cookieSessions: controlcookie.NewCookieStore(database, cookieSalt),
	}

	apiHandler := api.NewAPIHandler(logger, database, ret.cookieSessions, ingestKey, ingestTrigger, site, pages, email)
	apiHandler.RegisterRoutes(ret.mux)

	return ret
}
