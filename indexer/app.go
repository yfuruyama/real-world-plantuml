package indexer

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"

	"github.com/go-chi/chi"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

type IndexCreateRequestBody struct {
	Url string `json:"url"`
}

type GitHubContentResponse struct {
	Path    string `json:"path"`
	Sha     string `json:"sha"`
	Content string `json:"content"`
}

func init() {
	router := chi.NewRouter()

	router.Post("/indexer/create", func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)

		decoder := json.NewDecoder(r.Body)
		var body IndexCreateRequestBody
		if err := decoder.Decode(&body); err != nil {
			log.Warningf(ctx, "%s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		log.Infof(ctx, "url: %s", body.Url)

		re := regexp.MustCompile(`^https://github.com/([^/]+)/([^/]+)/blob/([^/]+)/(.+)$`)
		matched := re.FindStringSubmatch(body.Url)
		if len(matched) != 5 {
			log.Warningf(ctx, "invalid github url")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		owner := matched[1]
		repo := matched[2]
		hash := matched[3]
		path := matched[4]

		apiUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s", owner, repo, path, hash)
		log.Infof(ctx, "apiUrl: %s", apiUrl)

		token := os.Getenv("GITHUB_API_TOKEN")
		req, _ := http.NewRequest("GET", apiUrl, nil)
		req.Header.Add("Authorization", fmt.Sprintf("token %s", token))

		client := urlfetch.Client(ctx)
		resp, err := client.Do(req)
		if err != nil {
			log.Criticalf(ctx, "Failed to request to GitHub: err=%s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		decoder = json.NewDecoder(resp.Body)
		var ghcResp GitHubContentResponse
		if err := decoder.Decode(&ghcResp); err != nil {
			log.Criticalf(ctx, "Failed to parse response: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Infof(ctx, "Get content response: %#v", ghcResp)
		contentBytes, err := base64.StdEncoding.DecodeString(ghcResp.Content)
		if err != nil {
			log.Criticalf(ctx, "Failed to parse GitHub content: err=%s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		content := string(contentBytes)

		rendererBaseUrl := os.Getenv("RENDERER_BASE_URL")
		renderer := NewRenderer(ctx, rendererBaseUrl)

		syntaxCheckerBaseUrl := os.Getenv("SYNTAX_CHECKER_BASE_URL")
		syntaxChecker := NewSyntaxChecker(ctx, syntaxCheckerBaseUrl)

		indexer, err := NewIndexer(ctx, renderer, syntaxChecker, body.Url, content)
		if err != nil {
			log.Criticalf(ctx, "Failed to create indexer: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = indexer.Process()
		if err != nil {
			log.Criticalf(ctx, "%s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "ok")
	})

	router.Post("/_ah/push-handlers/gcs_notification", HandleGcsNotification)

	http.Handle("/", router)
}
