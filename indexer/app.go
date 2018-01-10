package indexer

import (
	"net/http"

	"github.com/go-chi/chi"
)

func init() {
	router := chi.NewRouter()

	router.Post("/index", HandleIndexCreate)
	router.Post("/_ah/push-handlers/gcs_notification", HandleGcsNotification)

	http.Handle("/", router)
}
