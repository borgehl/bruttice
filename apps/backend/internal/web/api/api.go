package api

import (
	"encoding/json"
	"github.com/borgehl/bruttice/apps/backend/internal/web/controlcookie"
	"github.com/borgehl/bruttice/apps/backend/internal/web/middleware"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/dbx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/emailx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/geoipx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/logx"
	"net/http"
)

// IngestTrigger starts a run on the scraper service.
type IngestTrigger interface {
	// returns false if a run is already in progress
	TryRunAll() (bool, error)
}

type APIHandler struct {
	logger         *logx.Logger
	database       *dbx.DatabaseClient
	email          emailx.EmailClient
	cookieSessions *controlcookie.CookieStore
	geoClient      *geoipx.GeoIpClient
	ingestKey      string
	ingestTrigger  IngestTrigger
}

func NewAPIHandler(logger *logx.Logger, database *dbx.DatabaseClient, cookieSessions *controlcookie.CookieStore, ingestKey string, ingestTrigger IngestTrigger) *APIHandler {
	return &APIHandler{
		logger:         logger,
		database:       database,
		email:          emailx.NewEmailClient(logger),
		cookieSessions: cookieSessions,
		geoClient:      geoipx.NewGeoIpClient(logger),
		ingestKey:      ingestKey,
		ingestTrigger:  ingestTrigger,
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
	mux.HandleFunc("GET /api/admin/ingest-runs", h.handleAdminIngestRuns)
	mux.HandleFunc("POST /api/admin/ingest-run", h.handleAdminStartIngest)
	mux.HandleFunc("POST /api/admin/vote", h.handleAdminVote)
	mux.HandleFunc("POST /api/admin/add-event", h.handleAdminAddEvent)
	mux.HandleFunc("PATCH /api/admin/event", h.handleAdminUpdateEvent)
	mux.HandleFunc("GET /api/admin/ages", h.handleAdminGetAges)

	// Internal ingest route — only reachable on the internal docker network
	mux.HandleFunc("POST /internal/ingest", h.handleIngest)
	mux.HandleFunc("POST /internal/ingest/runs", h.handleStartIngestRun)
	mux.HandleFunc("POST /internal/ingest/runs/{id}/finish", h.handleFinishIngestRun)
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
