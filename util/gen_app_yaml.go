package main

import (
	"flag"
	"log"
	"os"
	"strings"
	"text/template"
)

func main() {
	var inPath string
	var outPath string

	flag.StringVar(&inPath, "in", "", "Path to source app.yaml")
	flag.StringVar(&outPath, "out", "", "Path to target app.yaml")
	flag.Parse()

	if inPath == "" || outPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	tmpl, err := template.ParseFiles(inPath)
	if err != nil {
		log.Fatal(err)
	}

	out, err := os.Create(outPath)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	env := make(map[string]string)
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		env[pair[0]] = pair[1]
	}

	err = tmpl.Execute(out, env)
	if err != nil {
		log.Fatal(err)
	}
}
