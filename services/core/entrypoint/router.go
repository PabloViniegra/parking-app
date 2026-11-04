package entrypoint

import (
	"net/http"

	"github.com/PabloViniegra/parking-app/services/core/internal/bootstrap"
)

// The Vercel Go bridge compiles api as an external package, so api cannot import internal/bootstrap directly.
func NewRouter() http.Handler {
	return bootstrap.NewRouter()
}
