package ingesttrigger

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTryRunAll(t *testing.T) {
	cases := []struct {
		status  int
		started bool
		wantErr bool
	}{
		{http.StatusAccepted, true, false},
		{http.StatusConflict, false, false},
		{http.StatusUnauthorized, false, true},
		{http.StatusInternalServerError, false, true},
	}

	for _, tc := range cases {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer secret" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(tc.status)
		}))

		started, err := NewHTTPTrigger(srv.URL, "secret").TryRunAll()
		srv.Close()

		if started != tc.started || (err != nil) != tc.wantErr {
			t.Errorf("status %d: started=%v err=%v, want started=%v wantErr=%v", tc.status, started, err, tc.started, tc.wantErr)
		}
	}
}
