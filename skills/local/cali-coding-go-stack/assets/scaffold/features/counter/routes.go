package counter

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func SetupRoutes(r chi.Router) error {
	r.Get("/counter", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	})
	return nil
}
