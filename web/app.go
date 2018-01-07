package web

import (
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

func init() {
	gaTrackingId := os.Getenv("GA_TRACKING_ID")
	handler := NewHandler(gaTrackingId)

	router := chi.NewRouter()
	router.Get("/", handler.ToHandlerFunc(handler.GetIndex))
	router.Get("/search", handler.ToHandlerFunc(handler.GetSearch))
	router.NotFound(handler.ToHandlerFunc(handler.NotFound))
	// router.Get("/umls/{umlID:\\d+}", handler.ToHandlerFunc(handler.GetUml))

	http.Handle("/", router)
}
