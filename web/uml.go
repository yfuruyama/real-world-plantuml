package web

import (
	"context"
	"encoding/xml"
	"strconv"
	"strings"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/search"
)

type Uml struct {
	ID            int64       `datastore:"-"`
	GitHubUrl     string      `datastore:"gitHubUrl"`
	Source        string      `datastore:"source,noindex"`
	SourceSHA256  string      `datastore:"sourceSHA256"`
	DiagramType   DiagramType `datastore:"diagramType"`
	Svg           string      `datastore:"svg,noindex"`
	SvgViewBox    string      `datastore:"-"`
	PngBase64     string      `datastore:"pngBase64,noindex"`
	Ascii         string      `datastore:"ascii,noindex"`
	HighlightWord string      `datastore:"-"`
}

type SvgXml struct {
	ViewBox string `xml:"viewBox,attr"`
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

func (d DiagramType) ToHumanString() string {
	switch d {
	case TypeSequence:
		return "Sequence"
	case TypeUsecase:
		return "Use case"
	case TypeClass:
		return "Class"
	case TypeActivity:
		return "Activity"
	case TypeComponent:
		return "Component"
	case TypeState:
		return "State"
	}
	return ""
}

func FetchUmls(ctx context.Context, typ DiagramType, count int, cursor string) ([]*Uml, string, error) {
	q := datastore.NewQuery("Uml").Limit(count).KeysOnly()

	// Set filter
	if typ == TypeSequence || typ == TypeUsecase || typ == TypeClass || typ == TypeActivity || typ == TypeComponent || typ == TypeState {
		q = q.Filter("diagramType =", typ)
	}

	// Set cursor
	if cursor != "" {
		decoded, err := datastore.DecodeCursor(cursor)
		if err == nil {
			q = q.Start(decoded)
		}
	}

	// Do query
	iter := q.Run(ctx)
	var ids []int64
	for {
		key, err := iter.Next(nil)
		if err == datastore.Done {
			break
		}
		if err != nil {
			log.Criticalf(ctx, "datastore fetch error: %v", err)
			return nil, "", err
		}
		ids = append(ids, key.IntID())
	}

	umls, err := fetchUmlsByIds(ctx, ids)
	if err != nil {
		return nil, "", err
	}

	// Get nextCursor
	var nextCursor string
	if len(umls) == count {
		dsCursor, err := iter.Cursor()
		if err == nil {
			nextCursor = dsCursor.String()
		}
	}

	return umls, nextCursor, nil
}

func FetchUmlById(ctx context.Context, id int64) (*Uml, error) {
	umls, err := fetchUmlsByIds(ctx, []int64{id})
	if err != nil || len(umls) == 0 {
		return nil, err
	}
	return umls[0], nil
}

func SearchUmls(ctx context.Context, queryWord string, count int, cursor string) ([]*Uml, string, error) {
	fts, err := search.Open("uml_source")
	if err != nil {
		log.Criticalf(ctx, "failed to open FTS: %s", err)
		return nil, "", err
	}

	options := search.SearchOptions{
		Limit:   count,
		IDsOnly: true,
	}

	if cursor != "" {
		options.Cursor = search.Cursor(cursor)
	}

	query := strings.Join(strings.Split(queryWord, " "), " AND ")

	var ids []int64
	iter := fts.Search(ctx, query, &options)
	for {
		id, err := iter.Next(nil)
		if err == search.Done {
			break
		}
		if err != nil {
			log.Criticalf(ctx, "FTS search unexpected error: %v", err)
			break
		}
		intId, _ := strconv.ParseInt(id, 10, 64)
		ids = append(ids, intId)
	}
	log.Infof(ctx, "query result: %v", ids)

	var nextCursor string
	if len(ids) >= count {
		nextCursor = string(iter.Cursor())
	}

	umls, err := fetchUmlsByIds(ctx, ids)

	// for rendering
	for _, uml := range umls {
		uml.HighlightWord = queryWord
	}

	return umls, nextCursor, err
}

func fetchUmlsByIds(ctx context.Context, ids []int64) ([]*Uml, error) {
	keys := make([]*datastore.Key, len(ids))
	for i, id := range ids {
		keys[i] = datastore.NewKey(ctx, "Uml", "", id, nil)
	}
	umls := make([]*Uml, len(keys))
	notFounds := make([]bool, len(keys))

	err := datastore.GetMulti(ctx, keys, umls)
	if err != nil {
		multiErr, ok := err.(appengine.MultiError)
		if !ok {
			log.Criticalf(ctx, "Datastore fetch error: %v", err)
			return nil, err
		}
		for i, e := range multiErr {
			if e == nil {
				continue
			}
			if e == datastore.ErrNoSuchEntity {
				log.Warningf(ctx, "FTS index found, but datastore entity not found: %v", ids[i])
				notFounds[i] = true
				continue
			}
			log.Criticalf(ctx, "Datastore fetch partial error: %v", e)
			return nil, err
		}
	}

	var foundUmls []*Uml
	for i, notFound := range notFounds {
		if !notFound {
			uml := umls[i]
			uml.ID = ids[i]

			// Set viewBox
			var svgXml SvgXml
			err = xml.Unmarshal([]byte(uml.Svg), &svgXml)
			if err != nil {
				log.Criticalf(ctx, "svg parse error: %v", err)
			}
			uml.SvgViewBox = svgXml.ViewBox

			foundUmls = append(foundUmls, uml)
		}
	}

	return foundUmls, nil
}

func RegisterDummyUml(ctx context.Context) error {
	uml := &Uml{
		GitHubUrl: "https://github.com/yfuruyama/real-world-plantuml/blob/master/README.md",
		Source: `@startuml
Alice -> Bob: Authentication Request
Bob --> Alice: Authentication Response
@enduml`,
		SourceSHA256: "",
		DiagramType:  TypeSequence,
		Svg:          `<?xml version="1.0" encoding="UTF-8" standalone="no"?><svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" contentScriptType="application/ecmascript" contentStyleType="text/css" height="156px" preserveAspectRatio="none" style="width:246px;height:156px;" version="1.1" viewBox="0 0 246 156" width="246px" zoomAndPan="magnify"><defs><filter height="300%" id="fsv0g7ub4djg1" width="300%" x="-1" y="-1"><feGaussianBlur result="blurOut" stdDeviation="2.0"/><feColorMatrix in="blurOut" result="blurOut2" type="matrix" values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 .4 0"/><feOffset dx="4.0" dy="4.0" in="blurOut2" result="blurOut3"/><feBlend in="SourceGraphic" in2="blurOut3" mode="normal"/></filter></defs><g><line style="stroke: #A80036; stroke-width: 1.0; stroke-dasharray: 5.0,5.0;" x1="33" x2="33" y1="38.2969" y2="116.5625"/><line style="stroke: #A80036; stroke-width: 1.0; stroke-dasharray: 5.0,5.0;" x1="216" x2="216" y1="38.2969" y2="116.5625"/><rect fill="#FEFECE" filter="url(#fsv0g7ub4djg1)" height="30.2969" style="stroke: #A80036; stroke-width: 1.5;" width="46" x="8" y="3"/><text fill="#000000" font-family="sans-serif" font-size="14" lengthAdjust="spacingAndGlyphs" textLength="32" x="15" y="22.9951">Alice</text><rect fill="#FEFECE" filter="url(#fsv0g7ub4djg1)" height="30.2969" style="stroke: #A80036; stroke-width: 1.5;" width="46" x="8" y="115.5625"/><text fill="#000000" font-family="sans-serif" font-size="14" lengthAdjust="spacingAndGlyphs" textLength="32" x="15" y="135.5576">Alice</text><rect fill="#FEFECE" filter="url(#fsv0g7ub4djg1)" height="30.2969" style="stroke: #A80036; stroke-width: 1.5;" width="42" x="193" y="3"/><text fill="#000000" font-family="sans-serif" font-size="14" lengthAdjust="spacingAndGlyphs" textLength="28" x="200" y="22.9951">Bob</text><rect fill="#FEFECE" filter="url(#fsv0g7ub4djg1)" height="30.2969" style="stroke: #A80036; stroke-width: 1.5;" width="42" x="193" y="115.5625"/><text fill="#000000" font-family="sans-serif" font-size="14" lengthAdjust="spacingAndGlyphs" textLength="28" x="200" y="135.5576">Bob</text><polygon fill="#A80036" points="204,65.2969,214,69.2969,204,73.2969,208,69.2969" style="stroke: #A80036; stroke-width: 1.0;"/><line style="stroke: #A80036; stroke-width: 1.0;" x1="33" x2="210" y1="69.2969" y2="69.2969"/><text fill="#000000" font-family="sans-serif" font-size="13" lengthAdjust="spacingAndGlyphs" textLength="149" x="40" y="64.3638">Authentication Request</text><polygon fill="#A80036" points="44,94.4297,34,98.4297,44,102.4297,40,98.4297" style="stroke: #A80036; stroke-width: 1.0;"/><line style="stroke: #A80036; stroke-width: 1.0; stroke-dasharray: 2.0,2.0;" x1="38" x2="215" y1="98.4297" y2="98.4297"/><text fill="#000000" font-family="sans-serif" font-size="13" lengthAdjust="spacingAndGlyphs" textLength="159" x="50" y="93.4966">Authentication Response</text></g></svg>`,
		PngBase64:    `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`,
		Ascii: `┌─────┐                   ┌───┐
     │Alice│                   │Bob│
     └──┬──┘                   └─┬─┘
        │Authentication Request  │  
       └└—───────────────────>│  
        │                        │  
        │Authentication Response │  
        │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│  
     ┌──┴──┐                   ┌─┴─┐
     │Alice│                   │Bob│
     └─────┘                   └───┘
`,
	}

	key := datastore.NewIncompleteKey(ctx, "Uml", nil)
	_, err := datastore.Put(ctx, key, uml)
	return err
}
