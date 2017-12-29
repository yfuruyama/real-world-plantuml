package indexer

import (
	"encoding/base64"
	"fmt"
	"strings"
)

type Indexer struct {
	Content string
}

type Uml struct {
	Source string
	Type   DiagramType
	Svg    string
	Ascii  string
	Check  string
	ref    GitHubReference
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

// func guessDiagramType(source string) DiagramType {

// (N participants) => sequence
// (N entities) => usecase
// (N entities) => class
// (N activities) => activity
// (N entities) => component
// (N entities) => state
// (N entities) => object
// }

func NewIndexer(owner, repo, hash string, resp GitHubContentResponse) (*Indexer, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(resp.Content)
	if err != nil {
		return nil, err
	}

	content := string(contentBytes)
	return &Indexer{
		Content: content,
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
