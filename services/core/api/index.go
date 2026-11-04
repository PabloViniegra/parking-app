// Package handler is the single Vercel Functions entrypoint; it only delegates to the router.
package handler

import (
	"net/http"

	"github.com/PabloViniegra/parking-app/services/core/internal/bootstrap"
)

var router = bootstrap.NewRouter()

func Handler(w http.ResponseWriter, r *http.Request) {
	router.ServeHTTP(w, r)
}
