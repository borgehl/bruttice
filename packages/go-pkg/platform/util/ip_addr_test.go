package util

import (
	"net/http"
	"testing"
)

func TestGetClientIPAddr(t *testing.T) {
	tests := []struct {
		name    string
		request *http.Request
		want    string
	}{
		{
			name: "returns first IP from X-Forwarded-For",
			request: &http.Request{
				Header: http.Header{"X-Forwarded-For": []string{"1.2.3.4, 5.6.7.8"}},
			},
			want: "1.2.3.4",
		},
		{
			// http.Header map literals don't canonicalize keys — use Set instead
			name: "returns IP from X-Real-IP when no X-Forwarded-For",
			request: func() *http.Request {
				h := make(http.Header)
				h.Set("X-Real-IP", "9.9.9.9")
				return &http.Request{Header: h}
			}(),
			want: "9.9.9.9",
		},
		{
			name: "X-Forwarded-For takes priority over X-Real-IP",
			request: func() *http.Request {
				h := make(http.Header)
				h.Set("X-Forwarded-For", "1.2.3.4")
				h.Set("X-Real-IP", "9.9.9.9")
				return &http.Request{Header: h}
			}(),
			want: "1.2.3.4",
		},
		{
			name: "falls back to RemoteAddr when no proxy headers",
			request: &http.Request{
				Header:     http.Header{},
				RemoteAddr: "192.168.1.1:8080",
			},
			want: "192.168.1.1",
		},
		{
			name: "trims spaces from X-Forwarded-For value",
			request: &http.Request{
				Header: http.Header{"X-Forwarded-For": []string{"  1.2.3.4  , 5.6.7.8"}},
			},
			want: "1.2.3.4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetClientIPAddr(tt.request)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
