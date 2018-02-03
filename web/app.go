package web

import (
	"net/http"
	"os"

	"github.com/go-chi/chi"

	"google.golang.org/appengine"
)

func init() {
	gaTrackingId := os.Getenv("GA_TRACKING_ID")
	handler := NewHandler(gaTrackingId)

	router := chi.NewRouter()
	router.Get("/", handler.ToHandlerFunc(handler.GetIndex))
	router.Get("/search", handler.ToHandlerFunc(handler.GetSearch))
	router.Get("/umls/{umlID:\\d+}", handler.ToHandlerFunc(handler.GetUml))
	router.NotFound(handler.ToHandlerFunc(handler.NotFound))

	// for debugging
	if appengine.IsDevAppServer() {
		router.Get("/debug/dummy_uml", handler.ToHandlerFunc(handler.DebugRegisterDummyUml))
	}

	http.Handle("/", router)
}
