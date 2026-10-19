package main

import (
	"github.com/borgehl/bruttice/apps/backend/internal/daemons/notifications"
	"github.com/borgehl/bruttice/apps/backend/internal/web"
	"github.com/borgehl/bruttice/apps/backend/internal/web/api"
	"github.com/borgehl/bruttice/apps/backend/internal/web/ingesttrigger"
	"github.com/borgehl/bruttice/apps/backend/internal/web/mapper"
	"github.com/borgehl/bruttice/packages/go-pkg/configs"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/configx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/dbx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/logx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/passwordsx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/util"
	"net/http"

	_ "time/tzdata"
)

type Server struct {
	Mux      *http.ServeMux
	Logger   *logx.Logger
	Database *dbx.DatabaseClient

	Web *web.Controller
}

func (s Server) Run() {
	s.Logger.Info("starting web server on :8080")
	err := http.ListenAndServe(":8080", s.Mux)
	if err != nil {
		s.Logger.Panicf("server down: %s", err)
	}
	s.Logger.Info("Server shutdown")
}

func (s Server) HandleConfigs(cfg configx.Config) {
	// cfg holds the ingest key, cookie salt and admin passwords, so only log what is safe to
	s.Logger.Infof(
		"config loaded: timezone %s, %d admin accounts, %d tags, %d age groups, %d regions, %d venues",
		cfg.TZLocation, len(cfg.AdminAccounts),
		len(cfg.InitialData.Tags), len(cfg.InitialData.AgeGroups), len(cfg.InitialData.Regions), len(cfg.InitialData.Venues),
	)

	var err error
	tx := s.Database.DB.Begin()
	defer tx.Commit()

	for _, account := range cfg.AdminAccounts {
		s.Logger.Infof("adding admin email %s", account.Email)
		user := models.User{Email: account.Email}
		user, err = s.Database.GetUser(user, tx)
		if err != nil {
			tx.Rollback()
			panic(err)
		}
		// the root user is created without a password by Migrate, so an existing
		// user only counts as saved once it has one
		if user.ID != 0 && len(user.PasswordHash) > 0 {
			// already saved
			continue
		}
		user.PasswordHash, err = passwordsx.GetPasswordHash(account.Password)
		if err != nil {
			panic(err)
		}
		user, err = s.Database.SaveUser(user, tx)
		if err != nil {
			panic(err)
		}
	}
}

func main() {
	server := Server{
		Mux:    http.NewServeMux(),
		Logger: logx.NewLogger(nil, nil),
	}

	cfg := configx.ParseConfig(nil)
	util.SetTZLocation(cfg.TZLocation)
	if len(cfg.AdminAccounts) > 0 {
		configs.SetRootEmail(cfg.AdminAccounts[0].Email)
	}

	server.Database = dbx.NewDatabaseClient(server.Logger)
	server.Database.Migrate(cfg)

	server.HandleConfigs(cfg)

	// a nil interface (not a nil *HTTPTrigger) tells the API there is no scraper service
	var ingestTrigger api.IngestTrigger
	if cfg.IngestTriggerURL != "" {
		ingestTrigger = ingesttrigger.NewHTTPTrigger(cfg.IngestTriggerURL, cfg.IngestKey)
	}

	server.Web = web.NewController(server.Logger, server.Database, server.Mux, cfg.IngestKey, ingestTrigger, cfg.CookieSalt, mapper.SiteResponse(cfg.Site))

	go notifications.GoRunSendApprovalNotifications()

	server.Run()
}
