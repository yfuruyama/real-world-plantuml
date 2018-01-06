package web

import (
	"context"
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/search"
)

type Uml struct {
	ID          int64       `datastore:"-"`
	GitHubUrl   string      `datastore:"gitHubUrl"`
	Source      string      `datastore:"source,noindex"`
	DiagramType DiagramType `datastore:"diagramType"`
	Svg         string      `datastore:"svg,noindex"`
	SvgViewBox  string      `datastore:"-"`
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
)

type SvgXml struct {
	ViewBox string `xml:"viewBox,attr"`
}

type GlobalTemplateVars struct {
	GA_TRACKING_ID string
}

type FTSDocument struct {
	Document string `search:"document"`
}

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

		q := datastore.NewQuery("Uml").Limit(NUM_OF_ITEMS_PER_PAGE)

		// Set filter
		if typ == TypeSequence || typ == TypeUsecase || typ == TypeClass || typ == TypeActivity || typ == TypeComponent || typ == TypeState {
			q = q.Filter("diagramType =", typ)
		}

		// Set cursor
		if cursor := queryParams.Get("cursor"); cursor != "" {
			decoded, err := datastore.DecodeCursor(cursor)
			if err == nil {
				q = q.Start(decoded)
			}
		}

		// Do query
		iter := q.Run(ctx)
		var umls []Uml
		for {
			var uml Uml
			key, err := iter.Next(&uml)
			if err == datastore.Done {
				log.Infof(ctx, "iter done")
				break
			}
			if err != nil {
				log.Criticalf(ctx, "datastore fetch error: %v", err)
				break
			}
			uml.ID = key.IntID()

			// Set viewBox
			var svgXml SvgXml
			err = xml.Unmarshal([]byte(uml.Svg), &svgXml)
			if err != nil {
				log.Criticalf(ctx, "svg parse error: %v", err)
			}
			uml.SvgViewBox = svgXml.ViewBox

			umls = append(umls, uml)
		}

		// Get nextCursor
		var nextCursor string
		if len(umls) >= NUM_OF_ITEMS_PER_PAGE {
			dsCursor, err := iter.Cursor()
			if err == nil {
				nextCursor = dsCursor.String()
				log.Infof(ctx, "next cursor: %s", nextCursor)
			}
		}

		// TODO: マークアップが安定してきたら外に出す
		tmpl := template.Must(template.New("").Funcs(funcMap).ParseFiles("templates/base.html", "templates/index.html", "templates/components/uml_list.html"))

		err := tmpl.ExecuteTemplate(w, "base", UmlListTemplateVars{
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
		queryWord := r.URL.Query().Get("q")

		fts, err := search.Open("uml_source")
		if err != nil {
			log.Criticalf(ctx, "failed to open FTS: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		options := search.SearchOptions{
			Limit:   NUM_OF_ITEMS_PER_PAGE,
			IDsOnly: true,
		}

		cursor := r.URL.Query().Get("cursor")
		if cursor != "" {
			options.Cursor = search.Cursor(cursor)
		}

		query := fmt.Sprintf("document = \"%s\"", queryWord)

		var nextCursor string
		var entityIds []int64
		for iter := fts.Search(ctx, query, &options); ; {
			id, err := iter.Next(nil)
			if err == search.Done {
				if len(entityIds) >= NUM_OF_ITEMS_PER_PAGE {
					nextCursor = string(iter.Cursor())
				}
				break
			}
			if err != nil {
				log.Criticalf(ctx, "FTS search unexpected error: %v", err)
				break
			}
			intId, _ := strconv.ParseInt(id, 10, 64)
			entityIds = append(entityIds, intId)
		}
		log.Infof(ctx, "query result: %v", entityIds)

		keys := make([]*datastore.Key, len(entityIds))
		for i, id := range entityIds {
			keys[i] = datastore.NewKey(ctx, "Uml", "", id, nil)
		}
		umls := make([]Uml, len(keys))
		notFounds := make([]bool, len(keys))

		err = datastore.GetMulti(ctx, keys, umls)
		if err != nil {
			multiErr, ok := err.(appengine.MultiError)
			if !ok {
				log.Criticalf(ctx, "Datastore fetch error: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			for i, e := range multiErr {
				if e == nil {
					continue
				}
				if e == datastore.ErrNoSuchEntity {
					log.Warningf(ctx, "FTS index found, but datastore entity not found: %v", entityIds[i])
					notFounds[i] = true
					continue
				}
				log.Criticalf(ctx, "Datastore fetch partial error: %v", e)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		var filteredUmls []Uml
		for i, notFound := range notFounds {
			if !notFound {
				uml := umls[i]
				// Set viewBox
				var svgXml SvgXml
				err = xml.Unmarshal([]byte(uml.Svg), &svgXml)
				if err != nil {
					log.Criticalf(ctx, "svg parse error: %v", err)
				}
				uml.SvgViewBox = svgXml.ViewBox
				filteredUmls = append(filteredUmls, uml)
			}
		}

		// TODO: マークアップが安定してきたら外に出す
		tmpl := template.Must(template.New("").Funcs(funcMap).ParseFiles("templates/base.html", "templates/search.html", "templates/components/uml_list.html"))

		err = tmpl.ExecuteTemplate(w, "base", UmlListTemplateVars{
			CommonTemplateVars: &CommonTemplateVars{
				GATrackingID: gaTrackingId,
				Context:      ctx,
				Query:        queryWord,
			},
			Umls:       filteredUmls,
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
		tmpl := template.Must(template.New("").Funcs(funcMap).ParseFiles("templates/base.html", "templates/404.html"))
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
