package indexer

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type Indexer struct {
	Content  string
	Renderer *Renderer
	cxt      context.Context
}

type Uml struct {
	Source      string      `datastore:"source,noindex"`
	DiagramType DiagramType `datastore:"diagramType"`
	Svg         string      `datastore:"svg,noindex"`
	PngBase64   string      `datastore:"pngBase64,noindex"`
	Ascii       string      `datastore:"ascii,noindex"`
	// Check  string
	// ref    GitHubReference
}

type GitHubReference struct {
	Owner string
	Repo  string
	Path  string
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

func guessDiagramType(source, checked string) DiagramType {

	// (N participants) => sequence
	// (N entities) => usecase
	// (N entities) => class
	// (N activities) => activity
	// (N entities) => component
	// (N entities) => state
	// (N entities) => object
	return TypeUnknwon
}

func NewIndexer(cxt context.Context, renderer *Renderer, owner, repo, hash string, resp GitHubContentResponse) (*Indexer, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(resp.Content)
	if err != nil {
		return nil, err
	}

	content := string(contentBytes)
	return &Indexer{
		Content:  content,
		Renderer: renderer,
		cxt:      cxt,
	}, nil
}

func (idxr *Indexer) FindSources() []string {
	sources := make([]string, 0)
	content := idxr.Content
	for {
		startIdx := strings.Index(content, "@startuml")
		endIdx := strings.Index(content, "@enduml")
		if startIdx == -1 || endIdx == -1 {
			break
		}

		// TODO: 最小文字数チェック
		source := fmt.Sprintf("%s@enduml", content[startIdx:endIdx])
		sources = append(sources, source)

		content = content[(endIdx + len("@enduml")):]
	}
	return sources
}

func (idxr *Indexer) Process() error {
	cxt := idxr.cxt
	sources := idxr.FindSources()
	renderer := idxr.Renderer

	for _, source := range sources {
		log.Infof(cxt, "process source: %s", source)

		checked, err := renderer.CheckSyntax(source)
		if err != nil {
			log.Criticalf(cxt, "failed to check syntax: %s", err)
			return err
		}

		if checked == "(Error)" {
			log.Infof(cxt, "invalid syntax: %s", source)
			continue
		}

		typ := guessDiagramType(source, checked)

		svg, err := renderer.RenderSvg(source)
		if err != nil {
			log.Criticalf(cxt, "failed to render svg: %s", err)
			return err
		}

		png, err := renderer.RenderPng(source)
		if err != nil {
			log.Criticalf(cxt, "failed to render png: %s", err)
			return err
		}
		pngBase64 := base64.StdEncoding.EncodeToString(png)

		ascii, err := renderer.RenderAscii(source)
		if err != nil {
			log.Criticalf(cxt, "failed to render ascii: %s", err)
			return err
		}

		log.Infof(cxt, "make index: type=%s, svg=%s, pngBase64=%s, ascii=%s", typ, svg, pngBase64, ascii)

		uml := &Uml{
			Source:      source,
			DiagramType: typ,
			Svg:         svg,
			PngBase64:   pngBase64,
			Ascii:       ascii,
		}

		key := datastore.NewIncompleteKey(cxt, "Uml", nil)
		key, err = datastore.Put(cxt, key, uml)
		if err != nil {
			log.Criticalf(cxt, "put error: %s", err)
			return err
		}
	}

	return nil
}
