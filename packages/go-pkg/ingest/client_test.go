package ingest_test

import (
	"context"
	"encoding/json"
	"github.com/borgehl/bruttice/packages/go-pkg/ingest"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRunLifecycle(t *testing.T) {
	var paths []string
	var posted ingest.IngestionOutput
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer secret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		paths = append(paths, r.URL.Path)
		switch r.URL.Path {
		case "/internal/ingest/runs":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{"id": 7})
		case "/internal/ingest":
			json.NewDecoder(r.Body).Decode(&posted)
			json.NewEncoder(w).Encode(ingest.Result{Saved: 3})
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer srv.Close()

	c := ingest.NewClient(srv.URL, "secret")
	ctx := context.Background()

	runID, err := c.StartRun(ctx)
	if err != nil || runID != 7 {
		t.Fatalf("StartRun = %d, %v; want 7", runID, err)
	}

	result, err := c.Send(ctx, runID, ingest.IngestionOutput{VenueInfo: ingest.VenueInfo{Name: "Venue"}})
	if err != nil || result.Saved != 3 {
		t.Fatalf("Send = %+v, %v; want 3 saved", result, err)
	}
	if posted.RunID != 7 || posted.VenueInfo.Name != "Venue" {
		t.Errorf("server received %+v", posted)
	}

	if err := c.FinishRun(ctx, runID); err != nil {
		t.Fatal(err)
	}

	want := []string{"/internal/ingest/runs", "/internal/ingest", "/internal/ingest/runs/7/finish"}
	if len(paths) != len(want) {
		t.Fatalf("paths = %v, want %v", paths, want)
	}
	for i := range want {
		if paths[i] != want[i] {
			t.Errorf("paths[%d] = %s, want %s", i, paths[i], want[i])
		}
	}
}

func TestRejectedKeyIsAnError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	if _, err := ingest.NewClient(srv.URL, "wrong").StartRun(context.Background()); err == nil {
		t.Error("expected an error for a 401")
	}
}

func TestTrailingSlashInBaseURL(t *testing.T) {
	var path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	if err := ingest.NewClient(srv.URL+"/", "k").FinishRun(context.Background(), 3); err != nil {
		t.Fatal(err)
	}
	if path != "/internal/ingest/runs/3/finish" {
		t.Errorf("path = %s", path)
	}
}
