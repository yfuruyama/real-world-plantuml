package web

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-chi/chi"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

const NUM_OF_ITEMS_PER_PAGE = 10

type CommonTemplateVars struct {
	GATrackingID string
	Context      context.Context
	DiagramType  DiagramType
	Query        string
}

type UmlListTemplateVars struct {
	*CommonTemplateVars
	Umls       []*Uml
	NextCursor string
}

type Handler struct {
	GATrackingID string
	FuncMap      template.FuncMap
}

func (h *Handler) ToHandlerFunc(handle func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)
		err := handle(w, r)
		if err != nil {
			log.Criticalf(ctx, "%s", err)
			w.WriteHeader(http.StatusInternalServerError)

			// TODO: マークアップが安定してきたら外に出す
			tmpl := template.Must(template.New("").Funcs(h.FuncMap).ParseFiles(
				"templates/base.html",
				"templates/500.html",
			))
			_ = tmpl.ExecuteTemplate(w, "base", struct {
				*CommonTemplateVars
			}{
				CommonTemplateVars: &CommonTemplateVars{
					GATrackingID: h.GATrackingID,
					Context:      ctx,
				},
			})
		}
	})
}

func NewHandler(gaTrackingID string) *Handler {
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
			re, err := regexp.Compile(fmt.Sprintf("(?i)(%s)", word))
			if err != nil {
				return code
			}
			return re.ReplaceAllString(code, "<mark>$1</mark>")
		},
		"toUpperCase": func(word string) string {
			return strings.ToUpper(word)
		},
	}

	return &Handler{
		GATrackingID: gaTrackingID,
		FuncMap:      funcMap,
	}
}

func (h *Handler) GetIndex(w http.ResponseWriter, r *http.Request) error {
	ctx := appengine.NewContext(r)

	queryParams := r.URL.Query()
	typ := DiagramType(queryParams.Get("type"))
	cursor := queryParams.Get("cursor")

	umls, nextCursor, err := FetchUmls(ctx, typ, NUM_OF_ITEMS_PER_PAGE, cursor)
	if err != nil {
		return err
	}
	log.Debugf(ctx, "next cursor: %s", nextCursor)

	// TODO: マークアップが安定してきたら外に出す
	tmpl := template.Must(template.New("").Funcs(h.FuncMap).ParseFiles(
		"templates/base.html",
		"templates/index.html",
		"templates/components/uml_list.html",
		"templates/components/uml_item.html",
	))

	err = tmpl.ExecuteTemplate(w, "base", UmlListTemplateVars{
		CommonTemplateVars: &CommonTemplateVars{
			GATrackingID: h.GATrackingID,
			Context:      ctx,
			DiagramType:  typ,
		},
		Umls:       umls,
		NextCursor: nextCursor,
	})
	if err != nil {
		return err
	}

	return nil
}

func (h *Handler) GetSearch(w http.ResponseWriter, r *http.Request) error {
	ctx := appengine.NewContext(r)

	queryParams := r.URL.Query()
	query := queryParams.Get("q")
	cursor := queryParams.Get("cursor")

	if query == "" {
		return h.GetIndex(w, r)
	}

	umls, nextCursor, err := SearchUmls(ctx, query, NUM_OF_ITEMS_PER_PAGE, cursor)
	if err != nil {
		return err
	}
	log.Debugf(ctx, "next cursor: %s", nextCursor)

	// TODO: マークアップが安定してきたら外に出す
	tmpl := template.Must(template.New("").Funcs(h.FuncMap).ParseFiles(
		"templates/base.html",
		"templates/search.html",
		"templates/components/uml_list.html",
		"templates/components/uml_item.html",
	))

	err = tmpl.ExecuteTemplate(w, "base", UmlListTemplateVars{
		CommonTemplateVars: &CommonTemplateVars{
			GATrackingID: h.GATrackingID,
			Context:      ctx,
			Query:        query,
		},
		Umls:       umls,
		NextCursor: nextCursor,
	})
	if err != nil {
		return err
	}

	return nil
}

func (h *Handler) GetUml(w http.ResponseWriter, r *http.Request) error {
	ctx := appengine.NewContext(r)
	umlID, _ := strconv.ParseInt(chi.URLParam(r, "umlID"), 10, 64)

	uml, err := FetchUmlById(ctx, umlID)
	if err != nil {
		return err
	}

	if uml == nil {
		return h.NotFound(w, r)
	}

	// TODO: マークアップが安定してきたら外に出す
	tmpl := template.Must(template.New("").Funcs(h.FuncMap).ParseFiles(
		"templates/base.html",
		"templates/uml.html",
		"templates/components/uml_item.html",
	))

	err = tmpl.ExecuteTemplate(w, "base", struct {
		*CommonTemplateVars
		Uml Uml
	}{
		CommonTemplateVars: &CommonTemplateVars{
			GATrackingID: h.GATrackingID,
			Context:      ctx,
		},
		Uml: *uml,
	})
	if err != nil {
		return err
	}

	return nil
}

func (h *Handler) NotFound(w http.ResponseWriter, r *http.Request) error {
	ctx := appengine.NewContext(r)
	w.WriteHeader(http.StatusNotFound)

	// TODO: マークアップが安定してきたら外に出す
	tmpl := template.Must(template.New("").Funcs(h.FuncMap).ParseFiles(
		"templates/base.html",
		"templates/404.html",
	))
	_ = tmpl.ExecuteTemplate(w, "base", struct {
		*CommonTemplateVars
	}{
		CommonTemplateVars: &CommonTemplateVars{
			GATrackingID: h.GATrackingID,
			Context:      ctx,
		},
	})

	return nil
}
