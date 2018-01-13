package indexer

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/search"
)

const (
	MINIMUM_UML_SOURCE_LENGTH = 50
)

type Indexer struct {
	Renderer      *Renderer
	SyntaxChecker *SyntaxChecker
}

type Uml struct {
	GitHubUrl    string      `datastore:"gitHubUrl"`
	Source       string      `datastore:"source,noindex"`
	SourceSHA256 string      `datastore:"sourceSHA256"`
	DiagramType  DiagramType `datastore:"diagramType"`
	Svg          string      `datastore:"svg,noindex"`
	PngBase64    string      `datastore:"pngBase64,noindex"`
	Ascii        string      `datastore:"ascii,noindex"`
}

type DiagramType string

type FTSDocument struct {
	Document string `search:"document"`
}

const (
	TypeSequence  DiagramType = "sequence"
	TypeUsecase   DiagramType = "usecase"
	TypeClass     DiagramType = "class"
	TypeActivity  DiagramType = "activity"
	TypeComponent DiagramType = "component"
	TypeState     DiagramType = "state"
	// TypeObject    DiagramType = "object" // object is rarely detected
	TypeUnknwon DiagramType = "__unknown__"
)

func guessDiagramType(source string, result *SyntaxCheckResult) DiagramType {
	switch result.DiagramType {
	case "SEQUENCE":
		return TypeSequence
	case "DESCRIPTION":
		// Both of Usecase and Component diagram's syntax check results are "DESCRIPTION",
		// so distinct them ad hoc
		if strings.Contains(source, "actor") || strings.Contains(source, "usecase") {
			return TypeUsecase
		} else {
			return TypeComponent
		}
	case "CLASS":
		return TypeClass
	case "ACTIVITY":
		return TypeActivity
	case "STATE":
		return TypeState
	default:
		return TypeUnknwon
	}
}

func NewIndexer(renderer *Renderer, syntaxChecker *SyntaxChecker) *Indexer {
	return &Indexer{
		Renderer:      renderer,
		SyntaxChecker: syntaxChecker,
	}
}

func (idxr *Indexer) CreateIndexes(ctx context.Context, text string, gitHubUrl string) error {
	renderer := idxr.Renderer
	syntaxChecker := idxr.SyntaxChecker

	sources := findSources(ctx, text)
	for _, source := range sources {
		log.Infof(ctx, "process source: %s", source)
		if len(source) < MINIMUM_UML_SOURCE_LENGTH {
			log.Infof(ctx, "under minimum length: length=%d", len(source))
			continue
		}

		hash := sha256.Sum256([]byte(source))
		sourceHash := hex.EncodeToString(hash[:])
		log.Debugf(ctx, "source hash: %s", sourceHash)

		var existing []Uml
		q := datastore.NewQuery("Uml").Filter("sourceSHA256 =", sourceHash).Limit(1)
		_, err := q.GetAll(ctx, &existing)
		if err != nil {
			log.Criticalf(ctx, "failed to fetch existing umls: %v", err)
			return err
		}
		if len(existing) == 1 {
			log.Infof(ctx, "there is same uml existing: %#v", existing[0])
			continue
		}

		result, err := syntaxChecker.CheckSyntax(source)
		if err != nil {
			log.Criticalf(ctx, "failed to check syntax: %s", err)
			return err
		}
		log.Infof(ctx, "syntax check result: %v", result)

		if !result.Valid {
			log.Infof(ctx, "invalid syntax: %s", source)
			continue
		}
		if !result.HasValidDiagram() {
			log.Infof(ctx, "invalid diagram: %s", source)
			continue
		}

		typ := guessDiagramType(source, result)

		svg, err := renderer.RenderSvg(source)
		if err != nil {
			log.Criticalf(ctx, "failed to render svg: %s", err)
			return err
		}

		png, err := renderer.RenderPng(source)
		if err != nil {
			log.Criticalf(ctx, "failed to render png: %s", err)
			return err
		}
		pngBase64 := base64.StdEncoding.EncodeToString(png)

		ascii, err := renderer.RenderAscii(source)
		if err != nil {
			log.Criticalf(ctx, "failed to render ascii: %s", err)
			return err
		}

		log.Infof(ctx, "make index: type=%s, svg=%s, pngBase64=%s, ascii=%s", typ, svg, pngBase64, ascii)
		uml := &Uml{
			GitHubUrl:    gitHubUrl,
			Source:       source,
			SourceSHA256: sourceHash,
			DiagramType:  typ,
			Svg:          svg,
			PngBase64:    pngBase64,
			Ascii:        ascii,
		}

		key := datastore.NewIncompleteKey(ctx, "Uml", nil)
		key, err = datastore.Put(ctx, key, uml)
		if err != nil {
			log.Criticalf(ctx, "put error: %s", err)
			return err
		}

		// Register to full-text search index
		doc := FTSDocument{
			Document: source,
		}
		fts, err := search.Open("uml_source")
		if err != nil {
			log.Criticalf(ctx, "failed to open FTS: %s", err)
			return err
		}
		_, err = fts.Put(ctx, fmt.Sprintf("%d", key.IntID()), &doc)
		if err != nil {
			log.Criticalf(ctx, "failed to put document to FTS: %s", err)
			// Ignore error
			continue
		}
	}

	return nil
}

func findSources(ctx context.Context, text string) []string {
	sources := make([]string, 0)
	for {
		startIdx := strings.Index(text, "@startuml")
		endIdx := strings.Index(text, "@enduml")
		log.Debugf(ctx, "length:%d, startIdx:%d, endIdx:%d", len(text), startIdx, endIdx)
		if startIdx == -1 || endIdx == -1 {
			break
		}
		if startIdx < endIdx {
			source := fmt.Sprintf("%s@enduml", text[startIdx:endIdx])
			sources = append(sources, source)
		}

		text = text[(endIdx + len("@enduml")):]
	}
	return sources
}
