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

func (d DiagramType) String() string {
	switch d {
	case TypeSequence:
		return "Sequence"
	case TypeUsecase:
		return "Usecase"
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
