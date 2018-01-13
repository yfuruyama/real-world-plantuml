package indexer

import (
	"testing"

	"google.golang.org/appengine/aetest"
)

func TestFindSources(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	var tests = []struct {
		text     string
		expected []string
	}{
		{
			`
@startuml
alice -> bob
@enduml

@startuml
bob -> alice
@enduml
`, []string{"@startuml\nalice -> bob\n@enduml", "@startuml\nbob -> alice\n@enduml"}},
		{
			`
@startuml
alice -> bob
@enduml
@enduml
@startuml
`, []string{"@startuml\nalice -> bob\n@enduml"}},
		{
			`
@enduml
@startuml
alice -> bob
@enduml
@startuml
`, []string{"@startuml\nalice -> bob\n@enduml"}},
	}

	for _, test := range tests {
		got := findSources(ctx, test.text)
		if !isSameSources(got, test.expected) {
			t.Errorf("not expected sources: got=%#v, expected=%#v", got, test.expected)
		}
	}
}

func isSameSources(got []string, expected []string) bool {
	if len(got) != len(expected) {
		return false
	}
	for i := 0; i < len(got); i++ {
		if got[i] != expected[i] {
			return false
		}
	}
	return true
}
