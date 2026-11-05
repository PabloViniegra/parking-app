//go:build integration

package bootstrap_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/PabloViniegra/parking-app/services/core/internal/bootstrap"
)

func TestRouter_AuthenticatedAnnouncementPersistsAndReplays(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	var driver, key string
	if err := pool.QueryRow(ctx, `select gen_random_uuid()::text,gen_random_uuid()::text`).Scan(&driver, &key); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `insert into auth.users(id,aud,role,is_anonymous) values($1,'authenticated','authenticated',true)`, driver); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = pool.Exec(ctx, `delete from public.outbox where aggregate_id in (select id from public.spots where releaser_id=$1)`, driver)
		_, _ = pool.Exec(ctx, `delete from auth.users where id=$1`, driver)
	}()
	auth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/v1/user" || r.Header.Get("Authorization") != "Bearer test-access-token" || r.Header.Get("apikey") != "test-publishable-key" {
			http.Error(w, "Unauthorized", 401)
			return
		}
		_, _ = fmt.Fprintf(w, `{"id":%q,"role":"authenticated"}`, driver)
	}))
	defer auth.Close()
	t.Setenv("SUPABASE_URL", auth.URL)
	t.Setenv("SUPABASE_PUBLISHABLE_KEY", "test-publishable-key")
	router := bootstrap.NewRouter()
	submit := func(body string) *httptest.ResponseRecorder {
		request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/spots", strings.NewReader(body))
		request.Header.Set("Authorization", "Bearer test-access-token")
		request.Header.Set("Idempotency-Key", key)
		request.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		return recorder
	}
	first := submit(`{"latitude":40.4168,"longitude":-3.7038,"accuracyM":10}`)
	second := submit(`{"longitude":-3.7038,"accuracyM":10,"latitude":40.4168}`)
	if first.Code != 201 || second.Code != 201 || first.Body.String() != second.Body.String() {
		t.Fatalf("first=%d %s second=%d %s", first.Code, first.Body, second.Code, second.Body)
	}
	if conflict := submit(`{"latitude":40.4168,"longitude":-3.7038,"accuracyM":20}`); conflict.Code != 409 {
		t.Fatalf("conflict status=%d", conflict.Code)
	}
	var receipt struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(first.Body.Bytes(), &receipt); err != nil {
		t.Fatal(err)
	}
	var storedDriver string
	var confidence float64
	var events int
	if err := pool.QueryRow(ctx, `select releaser_id::text,confidence,(select count(*) from public.outbox where aggregate_id=s.id) from public.spots s where id=$1`, receipt.ID).Scan(&storedDriver, &confidence, &events); err != nil {
		t.Fatal(err)
	}
	if storedDriver != driver || confidence < 0.63999999 || confidence > 0.64000001 || events != 1 {
		t.Fatalf("driver=%s confidence=%f events=%d", storedDriver, confidence, events)
	}
}
