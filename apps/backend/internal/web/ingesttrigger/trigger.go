// Package ingesttrigger lets the web server ask the scraper service to start a run.
package ingesttrigger

import (
	"fmt"
	"net/http"
	"time"
)

// HTTPTrigger POSTs to the scraper service's trigger endpoint, authenticated with the ingest key.
type HTTPTrigger struct {
	url       string
	ingestKey string
	client    *http.Client
}

func NewHTTPTrigger(url string, ingestKey string) *HTTPTrigger {
	return &HTTPTrigger{
		url:       url,
		ingestKey: ingestKey,
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

// TryRunAll returns false if the scraper service reports a run is already in progress.
func (t *HTTPTrigger) TryRunAll() (bool, error) {
	req, err := http.NewRequest(http.MethodPost, t.url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+t.ingestKey)

	resp, err := t.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusAccepted:
		return true, nil
	case http.StatusConflict:
		return false, nil
	default:
		return false, fmt.Errorf("scraper service returned %s", resp.Status)
	}
}
