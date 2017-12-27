package main

import (
	"flag"
	"fmt"
)

type DiagramType string

const (
	TypeSequence  DiagramType = "sequence"
	TypeUsecase   DiagramType = "usecase"
	TypeClass     DiagramType = "class"
	TypeActivity  DiagramType = "activity"
	TypeComponent DiagramType = "component"
	TypeState     DiagramType = "state"
	TypeObject    DiagramType = "object"
)

func detectCategory(source string) DiagramType {

	// (N participants) => sequence
	// (N entities) => usecase
	// (N entities) => class
	// (N activities) => activity
	// (N entities) => component
	// (N entities) => state
	// (N entities) => object
}

func main() {
	var version string

	flag.StringVar(&version, "version", "", "Index version")
	flag.Parse()

	fmt.Println(version)

}
