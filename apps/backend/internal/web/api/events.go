package api

import (
	"errors"
	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
	"github.com/borgehl/bruttice/apps/backend/internal/web/mapper"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"net/http"
	"regexp"
	"time"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func (h *APIHandler) handleAddEvent(w http.ResponseWriter, r *http.Request) {
	var req dto.AddEventRequest
	if !h.decodeJSON(w, r, &req) {
		return
	}

	// Validate email
	if !emailRegex.MatchString(req.Email) {
		h.writeError(w, http.StatusBadRequest, "Invalid email address")
		return
	}
	// Validate required fields
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

	event, err := mapper.AddEventRequest(req)
	if err != nil {
		h.logger.Errorf("failed to map add event request: %v", err)
		h.writeError(w, http.StatusBadRequest, eventFormErrorMessage(err))
		return
	}

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

	// Send verification email asynchronously
	msg := h.email.VerifyEventMessage(event.Title, req.Email, *event.VerificationKey)
	go func() { _, _ = h.email.Send(msg) }()

	h.writeJSON(w, http.StatusOK, dto.AddEventResponse{
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

	// Update the event with the matching verification key
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

	h.writeJSON(w, http.StatusOK, dto.VerifyEventResponse{
		Success: true,
		Message: "Email verified successfully. Your event will be reviewed by an admin.",
	})
}

// eventFormErrorMessage is the message for an error from mapping an add event request.
func eventFormErrorMessage(err error) string {
	switch {
	case errors.Is(err, mapper.ErrInvalidDoorsTime):
		return "Invalid doors time format"
	default:
		return "Invalid date or start time format"
	}
}
