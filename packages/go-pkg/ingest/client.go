package ingest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client posts scraper output to a bruttice web server's /internal/ingest endpoints.
//
// It is the reference client for the ingest API, see docs/writing-scrapers.md.
// A client is safe for concurrent use.
type Client struct {
	baseURL string
	key     string

	// HTTP is the client used for requests. Replace it to change the timeout or transport.
	HTTP *http.Client
}

// Result is the server's response to a posted payload.
type Result struct {
	// Saved is the number of events the server saved.
	Saved int `json:"saved"`
	// Errors lists the events that failed to save. One bad event does not stop the rest.
	Errors []string `json:"errors"`
}

// NewClient returns a client for the web server at baseURL, for example
// "http://bruttice-backend:8080". The key is the server's ingest-key.
func NewClient(baseURL string, key string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		key:     key,
		HTTP:    &http.Client{Timeout: 2 * time.Minute},
	}
}

// StartRun opens a run on the server. Payloads sent with the returned id are
// grouped under it in the admin ingestion log.
func (c *Client) StartRun(ctx context.Context) (int64, error) {
	var resp struct {
		ID int64 `json:"id"`
	}
	if err := c.post(ctx, "/internal/ingest/runs", nil, &resp); err != nil {
		return 0, err
	}
	return resp.ID, nil
}

// FinishRun marks a run as finished. The server then emails the root admin any
// errors the run recorded.
func (c *Client) FinishRun(ctx context.Context, runID int64) error {
	return c.post(ctx, fmt.Sprintf("/internal/ingest/runs/%d/finish", runID), nil, nil)
}

// Send posts one venue's scraper output. A run id of 0 sends it outside any
// run, and the server records it as a run of its own.
func (c *Client) Send(ctx context.Context, runID int64, out IngestionOutput) (Result, error) {
	out.RunID = runID

	var result Result
	err := c.post(ctx, "/internal/ingest", out, &result)
	return result, err
}

func (c *Client) post(ctx context.Context, path string, body any, into any) error {
	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.key)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("POST %s: %s: %s", path, resp.Status, bytes.TrimSpace(msg))
	}
	if into == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(into)
}
