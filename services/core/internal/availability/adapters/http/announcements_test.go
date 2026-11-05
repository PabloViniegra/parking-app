package http_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	adapter "github.com/PabloViniegra/parking-app/services/core/internal/availability/adapters/http"
	"github.com/PabloViniegra/parking-app/services/core/internal/availability/application"
	availabilityDomain "github.com/PabloViniegra/parking-app/services/core/internal/availability/domain"
	availabilityPorts "github.com/PabloViniegra/parking-app/services/core/internal/availability/ports"
	coverageDomain "github.com/PabloViniegra/parking-app/services/core/internal/coverage/domain"
	coveragePorts "github.com/PabloViniegra/parking-app/services/core/internal/coverage/ports"
	"github.com/PabloViniegra/parking-app/services/core/internal/identity/adapters/supabase"
	"github.com/PabloViniegra/parking-app/services/core/internal/shared/clock"
)

func TestAnnouncements_RequireAuthentication(t *testing.T) {
	recorder := httptest.NewRecorder()
	adapter.NewAnnouncements(nil, nil).ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/spots", nil))
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", recorder.Code)
	}
}

type receiptRepository struct {
	spot    availabilityDomain.Spot
	event   availabilityDomain.Event
	receipt *availabilityPorts.Receipt
	saves   int
}

func (r *receiptRepository) Find(context.Context, string, string) (availabilityPorts.Receipt, error) {
	if r.receipt != nil {
		return *r.receipt, nil
	}
	return availabilityPorts.Receipt{}, availabilityPorts.ErrAnnouncementNotFound
}

func (r *receiptRepository) Save(_ context.Context, _ string, fingerprint string, spot availabilityDomain.Spot, event availabilityDomain.Event) (availabilityPorts.Receipt, error) {
	r.spot, r.event = spot, event
	r.saves++
	r.receipt = &availabilityPorts.Receipt{ID: spot.ID(), ExpiresAt: spot.TTLAt(), Fingerprint: fingerprint}
	return *r.receipt, nil
}

func TestAnnouncements_RetryPreservesOriginalResponseAndTime(t *testing.T) {
	repository := &receiptRepository{}
	clock := clock.NewFake(time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC))
	service := application.NewSubmitter(repository, locatedPolicy{}, clock, func() string { return "00000000-0000-4000-8000-000000000003" })
	handler := adapter.NewAnnouncements(&authenticator{}, service)
	submit := func() *httptest.ResponseRecorder {
		request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/spots", strings.NewReader(`{"latitude":40.4168,"longitude":-3.7038,"accuracyM":10}`))
		request.Header.Set("Authorization", "Bearer valid")
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Idempotency-Key", "00000000-0000-4000-8000-000000000004")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		return recorder
	}
	first := submit()
	clock.Advance(time.Minute)
	second := submit()
	if first.Code != 201 || second.Code != 201 || first.Body.String() != second.Body.String() || repository.saves != 1 {
		t.Fatalf("first=%s second=%s saves=%d", first.Body, second.Body, repository.saves)
	}
}

type locatedPolicy struct{}

func (locatedPolicy) Locate(context.Context, float64, float64) (coverageDomain.AnnouncementPolicy, error) {
	return coverageDomain.AnnouncementPolicy{CoverageID: "00000000-0000-4000-8000-000000000001", TTL: 180 * time.Second, Confidence: 0.64}, nil
}

func TestAnnouncements_DerivesIdentityTimeCellAndConfidenceOnServer(t *testing.T) {
	repository := &receiptRepository{}
	now := time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)
	service := application.NewSubmitter(repository, locatedPolicy{}, clock.NewFake(now), func() string { return "00000000-0000-4000-8000-000000000003" })
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/spots", strings.NewReader(`{"latitude":40.4168,"longitude":-3.7038,"accuracyM":10}`))
	request.Header.Set("Authorization", "Bearer valid")
	request.Header.Set("Idempotency-Key", "00000000-0000-4000-8000-000000000004")
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	adapter.NewAnnouncements(&authenticator{}, service).ServeHTTP(recorder, request)
	if recorder.Code != 201 {
		t.Fatalf("status = %d: %s", recorder.Code, recorder.Body)
	}
	if repository.spot.ReleaserID() != "00000000-0000-4000-8000-000000000002" || repository.spot.CellR10() != 622500375513628671 || repository.spot.Confidence() != 0.64 || !repository.spot.CreatedAt().Equal(now) || !repository.spot.TTLAt().Equal(time.Date(2026, 11, 5, 12, 3, 0, 0, time.UTC)) || repository.event.Type() != "availability.spot.announced.v1" {
		t.Fatalf("unexpected Spot: %+v", repository.spot)
	}
}

type transport func(*http.Request) (*http.Response, error)

func (f transport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestAnnouncements_VerifiesTokenWithSupabase(t *testing.T) {
	client := &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
		if r.URL.String() != "https://auth.example.test/auth/v1/user" || r.Header.Get("Authorization") != "Bearer signed-token" || r.Header.Get("apikey") != "publishable-test-key" {
			t.Fatalf("unexpected Auth request: %s", r.URL)
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"id":"00000000-0000-4000-8000-000000000002","role":"authenticated"}`)), Header: make(http.Header)}, nil
	})}
	service := &submitter{}
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/spots", strings.NewReader(`{"latitude":40,"longitude":-3,"accuracyM":10}`))
	request.Header.Set("Authorization", "Bearer signed-token")
	request.Header.Set("Idempotency-Key", "00000000-0000-4000-8000-000000000004")
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	adapter.NewAnnouncements(supabase.NewAuthenticator("https://auth.example.test", "publishable-test-key", client), service).ServeHTTP(recorder, request)
	if recorder.Code != 201 || service.driver != "00000000-0000-4000-8000-000000000002" {
		t.Fatalf("status = %d, driver = %q", recorder.Code, service.driver)
	}
}

func TestAnnouncements_DoNotSubmitWhenAuthFails(t *testing.T) {
	for _, tt := range []struct {
		name   string
		status int
		body   string
		want   int
	}{
		{name: "expired or invalid JWT", status: 401, body: `{}`, want: 401},
		{name: "Auth unavailable", status: 503, body: `{}`, want: 503},
		{name: "invalid Auth JSON", status: 200, body: `not json`, want: 503},
		{name: "privileged key is not a driver", status: 200, body: `{"id":"00000000-0000-4000-8000-000000000002","role":"service_role"}`, want: 401},
	} {
		t.Run(tt.name, func(t *testing.T) {
			client := &http.Client{Transport: transport(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: tt.status, Body: io.NopCloser(strings.NewReader(tt.body)), Header: make(http.Header)}, nil
			})}
			service := &submitter{}
			request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/spots", strings.NewReader(`{}`))
			request.Header.Set("Authorization", "Bearer invalid-token")
			recorder := httptest.NewRecorder()
			adapter.NewAnnouncements(supabase.NewAuthenticator("https://auth.example.test", "key", client), service).ServeHTTP(recorder, request)
			if recorder.Code != tt.want || service.driver != "" {
				t.Fatalf("status = %d, want %d; submitted=%t", recorder.Code, tt.want, service.driver != "")
			}
		})
	}
}

type authenticator struct{ token string }

func (a *authenticator) Verify(_ context.Context, token string) (string, error) {
	a.token = token
	return "00000000-0000-4000-8000-000000000002", nil
}

type submitter struct {
	driver, key string
	input       application.Submission
	fail        error
}

func (s *submitter) Submit(_ context.Context, driver, key string, input application.Submission) (application.Announcement, error) {
	s.driver, s.key, s.input = driver, key, input
	if s.fail != nil {
		return application.Announcement{}, s.fail
	}
	return application.Announcement{ID: "00000000-0000-4000-8000-000000000003", ExpiresAt: time.Date(2026, 11, 5, 12, 3, 0, 0, time.UTC)}, nil
}

func TestAnnouncements_MapsFailuresWithoutExposingDetails(t *testing.T) {
	for _, tt := range []struct {
		name string
		err  error
		want int
	}{
		{name: "idempotency conflict", err: availabilityPorts.ErrIdempotencyConflict, want: 409},
		{name: "outside Coverage", err: coveragePorts.ErrOutsideCoverage, want: 422},
		{name: "database failure", err: errors.New("database-password-and-coordinates"), want: 503},
	} {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/spots", strings.NewReader(`{"latitude":40,"longitude":-3,"accuracyM":10}`))
			request.Header.Set("Authorization", "Bearer valid")
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Idempotency-Key", "00000000-0000-4000-8000-000000000004")
			recorder := httptest.NewRecorder()
			adapter.NewAnnouncements(&authenticator{}, &submitter{fail: tt.err}).ServeHTTP(recorder, request)
			if recorder.Code != tt.want || strings.Contains(recorder.Body.String(), "database-password") {
				t.Fatalf("status=%d want=%d body=%s", recorder.Code, tt.want, recorder.Body)
			}
		})
	}
}

func TestAnnouncements_RejectsNonJSONMediaType(t *testing.T) {
	service := &submitter{}
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/spots", strings.NewReader(`{"latitude":40,"longitude":-3,"accuracyM":10}`))
	request.Header.Set("Authorization", "Bearer valid")
	request.Header.Set("Content-Type", "text/plain")
	request.Header.Set("Idempotency-Key", "00000000-0000-4000-8000-000000000004")
	recorder := httptest.NewRecorder()
	adapter.NewAnnouncements(&authenticator{}, service).ServeHTTP(recorder, request)
	if recorder.Code != 400 || service.driver != "" {
		t.Fatalf("status=%d submitted=%t", recorder.Code, service.driver != "")
	}
}

func TestAnnouncements_SubmitsForAuthenticatedReleaser(t *testing.T) {
	auth, service := &authenticator{}, &submitter{}
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/spots", strings.NewReader(`{"latitude":40.4168,"longitude":-3.7038,"accuracyM":10}`))
	request.Header.Set("Authorization", "Bearer valid-token")
	request.Header.Set("Idempotency-Key", "00000000-0000-4000-8000-000000000004")
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	adapter.NewAnnouncements(auth, service).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", recorder.Code, recorder.Body.String())
	}
	if auth.token != "valid-token" || service.driver != "00000000-0000-4000-8000-000000000002" || service.key != "00000000-0000-4000-8000-000000000004" || service.input != (application.Submission{Latitude: 40.4168, Longitude: -3.7038, AccuracyM: 10}) {
		t.Fatalf("unexpected submission: auth=%+v service=%+v", auth, service)
	}
	var response map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if len(response) != 2 || response["id"] != "00000000-0000-4000-8000-000000000003" || response["expiresAt"] != "2026-11-05T12:03:00Z" {
		t.Fatalf("response = %v", response)
	}
}

func TestAnnouncements_RejectsInvalidInputBeforeSubmission(t *testing.T) {
	for _, tt := range []struct{ name, body, key string }{
		{name: "missing latitude", body: `{"longitude":-3.7,"accuracyM":10}`, key: "00000000-0000-4000-8000-000000000004"},
		{name: "client identity", body: `{"latitude":40,"longitude":-3,"accuracyM":10,"releaserId":"other"}`, key: "00000000-0000-4000-8000-000000000004"},
		{name: "out of range", body: `{"latitude":91,"longitude":-3,"accuracyM":10}`, key: "00000000-0000-4000-8000-000000000004"},
		{name: "missing key", body: `{"latitude":40,"longitude":-3,"accuracyM":10}`},
		{name: "invalid accuracy", body: `{"latitude":40,"longitude":-3,"accuracyM":0}`, key: "00000000-0000-4000-8000-000000000004"},
		{name: "trailing JSON", body: `{"latitude":40,"longitude":-3,"accuracyM":10} {}`, key: "00000000-0000-4000-8000-000000000004"},
		{name: "oversize", body: strings.Repeat(" ", 4096) + `{"latitude":40,"longitude":-3,"accuracyM":10}`, key: "00000000-0000-4000-8000-000000000004"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			service := &submitter{}
			request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/spots", strings.NewReader(tt.body))
			request.Header.Set("Authorization", "Bearer valid-token")
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Idempotency-Key", tt.key)
			recorder := httptest.NewRecorder()
			adapter.NewAnnouncements(&authenticator{}, service).ServeHTTP(recorder, request)
			if recorder.Code != http.StatusBadRequest || service.driver != "" {
				t.Fatalf("status = %d, submitted = %t", recorder.Code, service.driver != "")
			}
		})
	}
}
