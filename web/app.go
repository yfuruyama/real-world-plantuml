package web

import (
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

// TODO: indexer と共有する
type Uml struct {
	GitHubUrl   string      `datastore:"gitHubUrl"`
	Source      string      `datastore:"source,noindex"`
	DiagramType DiagramType `datastore:"diagramType"`
	Svg         string      `datastore:"svg,noindex"`
	PngBase64   string      `datastore:"pngBase64,noindex"`
	Ascii       string      `datastore:"ascii,noindex"`
}
type DiagramType string

const (
	TypeSequence  DiagramType = "sequence"
	TypeUsecase   DiagramType = "usecase"
	TypeClass     DiagramType = "class"
	TypeActivity  DiagramType = "activity"
	TypeComponent DiagramType = "component"
	TypeState     DiagramType = "state"
	TypeObject    DiagramType = "object"
	TypeUnknwon   DiagramType = "__unknown__"
)

func init() {
	router := chi.NewRouter()
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)

		queryParams := r.URL.Query()
		typ := DiagramType(queryParams.Get("type"))

		var umls []Uml
		q := datastore.NewQuery("Uml").Limit(10)
		if typ == TypeSequence || typ == TypeUsecase || typ == TypeClass || typ == TypeActivity || typ == TypeComponent || typ == TypeState || typ == TypeObject {
			q = q.Filter("diagramType =", typ)
		}
		_, err := q.GetAll(ctx, &umls)

		// TODO: マークアップが安定してきたら外に出す
		tmpl := template.Must(template.New("").Funcs(template.FuncMap{
			"safehtml": func(text string) template.HTML {
				return template.HTML(text)
			},
			"loopLineTimes": func(text string) []struct{} {
				return make([]struct{}, strings.Count(text, "\n")+1)
			},
			"githubUrlToAnchorText": func(url string) string {
				re := regexp.MustCompile(`^https://github.com/([^/]+)/([^/]+)/(.+)/(.+)$`)
				matched := re.FindStringSubmatch(url)
				if len(matched) != 5 {
					log.Warningf(ctx, "invalid github url: %s", url)
					return ""
				}

				owner := matched[1]
				repo := matched[2]
				_ = matched[3]
				file := matched[4]
				return fmt.Sprintf("%s/%s - %s", owner, repo, file)
			},
		}).ParseFiles("templates/index.html"))

		err = tmpl.ExecuteTemplate(w, "index.html", struct {
			Umls []Uml
		}{
			umls,
		})
		if err != nil {
			log.Criticalf(ctx, "%s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
	http.Handle("/", router)
}
