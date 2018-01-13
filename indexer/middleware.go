package indexer

import (
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

func AuthTaskqueue(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)
		if r.Header.Get("X-AppEngine-QueueName") == "" {
			log.Warningf(ctx, "Request is not from TaskQueue")
			w.WriteHeader(http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
