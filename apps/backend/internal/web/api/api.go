package api

import (
	"github.com/borgehl/bruttice/apps/backend/internal/web/controlcookie"
	"github.com/borgehl/bruttice/apps/backend/internal/web/middleware"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/dbx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/emailx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/geoipx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/logx"
	"encoding/json"
	"net/http"
)

type APIHandler struct {
	logger         *logx.Logger
	database       *dbx.DatabaseClient
	email          emailx.EmailClient
	cookieSessions *controlcookie.CookieStore
	geoClient      *geoipx.GeoIpClient
	ingestKey      string
}

func NewAPIHandler(logger *logx.Logger, database *dbx.DatabaseClient, cookieSessions *controlcookie.CookieStore, ingestKey string) *APIHandler {
	return &APIHandler{
		logger:         logger,
		database:       database,
		email:          emailx.NewEmailClient(logger),
		cookieSessions: cookieSessions,
		geoClient:      geoipx.NewGeoIpClient(logger),
		ingestKey:      ingestKey,
	}
}

func (h *APIHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/calendar", middleware.MiddlewareVisitorTracker(h.database, h.logger, h.geoClient, h.handleGetCalendar))
	mux.HandleFunc("GET /api/filter-options", middleware.MiddlewareVisitorTracker(h.database, h.logger, h.geoClient, h.handleGetFilterOptions))
	mux.HandleFunc("GET /api/filter", h.handleGetFilter)
	mux.HandleFunc("GET /api/venues", h.handleGetVenues)
	mux.HandleFunc("GET /api/submission-venues", h.handleGetSubmissionVenues)
	mux.HandleFunc("POST /api/add-event", middleware.MiddlewareVisitorTracker(h.database, h.logger, h.geoClient, h.handleAddEvent))
	mux.HandleFunc("POST /api/visitor/consent", h.handleVisitorConsent)
	mux.HandleFunc("POST /api/filter", h.handleCreateFilter)
	mux.HandleFunc("GET /api/verify-event", h.handleVerifyEvent)
	mux.HandleFunc("GET /api/feeds", h.handleGetFeeds)

	// Admin routes
	mux.HandleFunc("POST /api/admin/login", h.handleAdminLogin)
	mux.HandleFunc("POST /api/admin/logout", h.handleAdminLogout)
	mux.HandleFunc("GET /api/admin/session", h.handleAdminCheckSession)
	mux.HandleFunc("GET /api/admin/dashboard", h.handleAdminDashboard)
	mux.HandleFunc("POST /api/admin/vote", h.handleAdminVote)
	mux.HandleFunc("POST /api/admin/add-event", h.handleAdminAddEvent)
	mux.HandleFunc("PATCH /api/admin/event", h.handleAdminUpdateEvent)
	mux.HandleFunc("GET /api/admin/ages", h.handleAdminGetAges)

	// Internal ingest route — only reachable on the internal docker network
	mux.HandleFunc("POST /internal/ingest", h.handleIngest)
}

func (h *APIHandler) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Errorf("failed to encode JSON response: %v", err)
	}
}

func (h *APIHandler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]string{"error": message})
}
