package web

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
)

func init() {
	router := chi.NewRouter()
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello")
	})
	http.Handle("/", router)
}
