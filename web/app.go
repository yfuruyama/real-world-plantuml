package web

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/go-chi/chi"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

type CommonTemplateVars struct {
	GATrackingID string
	Context      context.Context
	DiagramType  DiagramType
	Query        string
}

type UmlListTemplateVars struct {
	*CommonTemplateVars
	Umls       []Uml
	NextCursor string
}

const NUM_OF_ITEMS_PER_PAGE = 10

func init() {
	gaTrackingId := os.Getenv("GA_TRACKING_ID")

	funcMap := template.FuncMap{
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
				return ""
			}

			owner := matched[1]
			repo := matched[2]
			_ = matched[3]
			file := matched[4]
			text := fmt.Sprintf("%s/%s - %s", owner, repo, file)
			// abbreviation
			if len(text) > 40 {
				text = fmt.Sprintf("%s...%s", text[0:20], text[len(text)-20:len(text)])
			}
			return text
		},
		"staticPath": func(ctx context.Context, filePath string) string {
			return fmt.Sprintf("/static/%s?v=%s", filePath, appengine.VersionID(ctx))
		},
		"highlight": func(word, code string) string {
			if word == "" {
				return code
			}
			re, err := regexp.Compile(word)
			if err != nil {
				return code
			}
			return re.ReplaceAllString(code, fmt.Sprintf("<mark>%s</mark>", word))
		},
	}

	router := chi.NewRouter()
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)

		queryParams := r.URL.Query()
		typ := DiagramType(queryParams.Get("type"))
		cursor := queryParams.Get("cursor")

		umls, nextCursor, err := FetchUmls(ctx, typ, NUM_OF_ITEMS_PER_PAGE, cursor)
		if err != nil {
			log.Criticalf(ctx, "%s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Debugf(ctx, "next cursor: %s", nextCursor)

		// TODO: マークアップが安定してきたら外に出す
		tmpl := template.Must(template.New("").Funcs(funcMap).ParseFiles(
			"templates/base.html",
			"templates/index.html",
			"templates/components/uml_list.html",
		))

		err = tmpl.ExecuteTemplate(w, "base", UmlListTemplateVars{
			CommonTemplateVars: &CommonTemplateVars{
				GATrackingID: gaTrackingId,
				Context:      ctx,
				DiagramType:  typ,
			},
			Umls:       umls,
			NextCursor: nextCursor,
		})
		if err != nil {
			log.Criticalf(ctx, "%s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	// router.Get("/umls/{umlID:\\d+}", func(w http.ResponseWriter, r *http.Request) {
	// ctx := appengine.NewContext(r)
	// umlID, _ := strconv.ParseInt(chi.URLParam(r, "umlID"), 10, 64)
	// key := datastore.NewKey(ctx, "Uml", "", umlID, nil)

	// var uml Uml
	// err := datastore.Get(ctx, key, &uml)
	// if err != nil {
	// if err == datastore.ErrNoSuchEntity {
	// log.Warningf(ctx, "Uml not found: %v", umlID)
	// handle404(w, r)
	// return
	// }

	// log.Criticalf(ctx, "Unexpected datastore error: %s", err)
	// w.WriteHeader(http.StatusInternalServerError)
	// return
	// }

	// // TODO: マークアップが安定してきたら外に出す
	// tmpl := template.Must(template.New("").Funcs(funcMap).ParseFiles("templates/base.html", "templates/uml.html"))

	// err = tmpl.ExecuteTemplate(w, "base", struct {
	// *GlobalTemplateVars
	// Uml Uml
	// }{
	// &globalTemplateVars,
	// uml,
	// })
	// if err != nil {
	// log.Criticalf(ctx, "%s", err)
	// w.WriteHeader(http.StatusInternalServerError)
	// return
	// }
	// })

	router.Get("/search", func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)

		queryParams := r.URL.Query()
		query := queryParams.Get("q")
		cursor := queryParams.Get("cursor")

		umls, nextCursor, err := SearchUmls(ctx, query, NUM_OF_ITEMS_PER_PAGE, cursor)
		if err != nil {
			log.Criticalf(ctx, "%s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Debugf(ctx, "next cursor: %s", nextCursor)

		// TODO: マークアップが安定してきたら外に出す
		tmpl := template.Must(template.New("").Funcs(funcMap).ParseFiles(
			"templates/base.html",
			"templates/search.html",
			"templates/components/uml_list.html",
		))

		err = tmpl.ExecuteTemplate(w, "base", UmlListTemplateVars{
			CommonTemplateVars: &CommonTemplateVars{
				GATrackingID: gaTrackingId,
				Context:      ctx,
				Query:        query,
			},
			Umls:       umls,
			NextCursor: nextCursor,
		})
		if err != nil {
			log.Criticalf(ctx, "%s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)
		w.WriteHeader(http.StatusNotFound)

		// TODO: マークアップが安定してきたら外に出す
		tmpl := template.Must(template.New("").Funcs(funcMap).ParseFiles(
			"templates/base.html",
			"templates/404.html",
		))
		_ = tmpl.ExecuteTemplate(w, "base", struct {
			*CommonTemplateVars
		}{
			CommonTemplateVars: &CommonTemplateVars{
				GATrackingID: gaTrackingId,
				Context:      ctx,
			},
		})
		return
	})

	http.Handle("/", router)
}
