package indexer

import (
	"net/http"

	"github.com/go-chi/chi"
)

func init() {
	router := chi.NewRouter()

	router.Route("/indexes", func(r chi.Router) {
		r.Use(AuthTaskqueue)
		r.Post("/", HandleIndexCreate)
	})
	router.Post("/_ah/push-handlers/gcs_notification", HandleGcsNotification)

	http.Handle("/", router)
}
