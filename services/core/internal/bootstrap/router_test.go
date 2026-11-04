package bootstrap_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PabloViniegra/parking-app/services/core/internal/bootstrap"
)

func TestRouter(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		target     string
		wantStatus int
		wantBody   string
	}{
		{name: "health check responds ok", method: http.MethodGet, target: "/healthz", wantStatus: http.StatusOK, wantBody: `{"status":"ok"}` + "\n"},
		{name: "health check rejects other methods", method: http.MethodPost, target: "/healthz", wantStatus: http.StatusMethodNotAllowed},
		{name: "unknown route is not found", method: http.MethodGet, target: "/v1/unknown", wantStatus: http.StatusNotFound},
	}

	router := bootstrap.NewRouter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), tt.method, tt.target, nil))

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && rec.Body.String() != tt.wantBody {
				t.Errorf("body = %q, want %q", rec.Body.String(), tt.wantBody)
			}
		})
	}
}
