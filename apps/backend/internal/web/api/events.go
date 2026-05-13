package api

import (
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/util"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const formDTFormat = "2006-01-02 15:04"

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func normalizeURL(url string) string {
	url = strings.TrimSpace(url)
	if url == "" {
		return url
	}
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}
	return "https://" + url
}

func (h *APIHandler) handleAddEvent(w http.ResponseWriter, r *http.Request) {
	var req AddEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Errorf("failed to decode request body: %v", err)
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if !emailRegex.MatchString(req.Email) {
		h.writeError(w, http.StatusBadRequest, "Invalid email address")
		return
	}
	if req.Title == "" {
		h.writeError(w, http.StatusBadRequest, "Title is required")
		return
	}
	if req.Link == "" {
		h.writeError(w, http.StatusBadRequest, "Event URL is required")
		return
	}
	if req.Venue == "" {
		h.writeError(w, http.StatusBadRequest, "Venue is required")
		return
	}
	if req.Date == "" {
		h.writeError(w, http.StatusBadRequest, "Date is required")
		return
	}
	if req.StartTime == "" {
		h.writeError(w, http.StatusBadRequest, "Start time is required")
		return
	}

	var event models.Event
	event.VerificationEmail = &req.Email
	event.Title = req.Title
	event.Link = normalizeURL(req.Link)
	event.CoverCost = req.CoverCost
	event.Venue.Name = req.Venue
	event.Region.Name = req.Region
	event.Age.Name = req.Age

	startAt, err := time.ParseInLocation(formDTFormat, fmt.Sprintf("%s %s", req.Date, req.StartTime), util.TZLocation)
	if err != nil {
		h.logger.Errorf("failed to parse start time: %v", err)
		h.writeError(w, http.StatusBadRequest, "Invalid date or start time format")
		return
	}
	event.StartAt = startAt

	if req.DoorsTime != "" {
		doorsAt, err := time.ParseInLocation(formDTFormat, fmt.Sprintf("%s %s", req.Date, req.DoorsTime), util.TZLocation)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, "Invalid doors time format")
			return
		}
		event.DoorsAt = &doorsAt
	}

	for _, tag := range req.Tags {
		event.Tags = append(event.Tags, models.Tag{Name: tag})
	}

	event.CreateVerificationKey()
	event.UpdateGuid(nil)

	tx := h.database.DB.Begin()
	if tx.Error != nil {
		h.logger.Errorf("failed to begin transaction: %v", tx.Error)
		h.writeError(w, http.StatusInternalServerError, "Failed to save event")
		return
	}
	defer func() {
		if err := tx.Commit().Error; err != nil {
			h.logger.Errorf("failed to commit transaction: %v", err)
		}
	}()

	event, err = h.database.SaveEvent(event, false, tx)
	if err != nil {
		h.logger.Errorf("failed to save event: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to save event")
		return
	}

	msg := h.email.GetVerifyEventEmail(event.Title, req.Email, *event.VerificationKey)
	go h.email.SendEmail(req.Email, msg)

	h.writeJSON(w, http.StatusOK, AddEventResponse{
		Success: true,
		Message: "Event submitted successfully. Please check your email to verify the submission.",
	})
}

func (h *APIHandler) handleVerifyEvent(w http.ResponseWriter, r *http.Request) {
	verifyKey := r.URL.Query().Get("verifyKey")
	if verifyKey == "" {
		h.writeError(w, http.StatusBadRequest, "Verification key is required")
		return
	}

	tx := h.database.DB.Begin()
	defer tx.Commit()

	result := tx.Model(&models.Event{}).
		Where("verification_key = ?", verifyKey).
		Update("verified_at", time.Now().Unix())

	if result.Error != nil {
		h.logger.Errorf("failed to verify event with key %s: %v", verifyKey, result.Error)
		h.writeError(w, http.StatusInternalServerError, "Failed to verify event")
		return
	}
	if result.RowsAffected == 0 {
		h.writeError(w, http.StatusNotFound, "Invalid or expired verification key")
		return
	}

	h.writeJSON(w, http.StatusOK, VerifyEventResponse{
		Success: true,
		Message: "Email verified successfully. Your event will be reviewed by an admin.",
	})
}
