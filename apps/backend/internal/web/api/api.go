package api

import (
	"encoding/json"
	"github.com/borgehl/bruttice/apps/backend/internal/web/controlcookie"
	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
	"github.com/borgehl/bruttice/apps/backend/internal/web/middleware"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/dbx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/emailx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/geoipx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/logx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/pagesx"
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
	email          *emailx.EmailClient
	cookieSessions *controlcookie.CookieStore
	geoClient      *geoipx.GeoIpClient
	ingestKey      string
	ingestTrigger  IngestTrigger
	site           dto.SiteResponse
	pages          []pagesx.Page
	contactLimit   *rateLimiter
	inbox          inboxStore
	authenticate   func(*http.Request) *models.User
	users          userStore
}

func NewAPIHandler(logger *logx.Logger, database *dbx.DatabaseClient, cookieSessions *controlcookie.CookieStore, ingestKey string, ingestTrigger IngestTrigger, site dto.SiteResponse, pages []pagesx.Page, email *emailx.EmailClient) *APIHandler {
	return &APIHandler{
		logger:         logger,
		database:       database,
		email:          email,
		cookieSessions: cookieSessions,
		geoClient:      geoipx.NewGeoIpClient(logger),
		ingestKey:      ingestKey,
		ingestTrigger:  ingestTrigger,
		site:           site,
		pages:          pages,
		contactLimit:   newContactLimiter(),
		inbox:          database,
		users:          database,
	}
}

func (h *APIHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/calendar", middleware.MiddlewareVisitorTracker(h.database, h.logger, h.geoClient, h.handleGetCalendar))
	mux.HandleFunc("GET /api/filter-options", middleware.MiddlewareVisitorTracker(h.database, h.logger, h.geoClient, h.handleGetFilterOptions))
	mux.HandleFunc("GET /api/filter", h.handleGetFilter)
	mux.HandleFunc("GET /api/site", h.handleGetSite)
	mux.HandleFunc("GET /api/pages", h.handleGetPages)
	mux.HandleFunc("GET /api/pages/{slug}", h.handleGetPage)
	mux.HandleFunc("GET /api/venues", h.handleGetVenues)
	mux.HandleFunc("GET /api/submission-venues", h.handleGetSubmissionVenues)
	mux.HandleFunc("POST /api/add-event", middleware.MiddlewareVisitorTracker(h.database, h.logger, h.geoClient, h.handleAddEvent))
	mux.HandleFunc("POST /api/contact", h.handleContact)
	mux.HandleFunc("POST /api/webhooks/resend", h.handleResendWebhook)
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
	mux.HandleFunc("PATCH /api/admin/settings", h.handleAdminSettings)
	mux.HandleFunc("GET /api/admin/inbox", h.handleAdminInbox)
	mux.HandleFunc("POST /api/admin/inbox/{threadId}/reply", h.handleAdminInboxReply)
	mux.HandleFunc("POST /api/admin/inbox/{threadId}/read", h.handleAdminInboxRead)
	mux.HandleFunc("POST /api/admin/inbox/{threadId}/archive", h.handleAdminInboxArchive(true))
	mux.HandleFunc("POST /api/admin/inbox/{threadId}/unarchive", h.handleAdminInboxArchive(false))

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
	h.writeJSON(w, status, dto.ErrorResponse{Error: message})
}

// maxRequestBody limits the body of a request. Ingest payloads set their own limit.
const maxRequestBody = 1 << 20

// decodeJSON reads the request body into v. If it can't, it writes a 400 and returns false.
func (h *APIHandler) decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBody)
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		h.logger.Errorf("%s %s: failed to decode request body: %v", r.Method, r.URL.Path, err)
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return false
	}
	return true
}

// decodeStrictJSON is decodeJSON for clients that are written against a contract, which
// refuses fields it doesn't know and says which one. limit is the largest body in bytes.
func (h *APIHandler) decodeStrictJSON(w http.ResponseWriter, r *http.Request, v any, limit int64) bool {
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(v); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return false
	}
	return true
}
