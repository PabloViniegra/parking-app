// Package handler is the single Vercel Functions entrypoint; it only delegates to the router.
package handler

import (
	"net/http"

	"github.com/PabloViniegra/parking-app/services/core/entrypoint"
)

var router = entrypoint.NewRouter()

func Handler(w http.ResponseWriter, r *http.Request) {
	router.ServeHTTP(w, r)
}
