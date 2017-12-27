package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	// "regexp"
	"io/ioutil"
)

func main() {
	var port int

	flag.IntVar(&port, "port", 8080, "Port for server")
	flag.Parse()

	source := os.Args[1]

	baseUrl := fmt.Sprintf("http://localhost:%d", port)

	values := url.Values{}
	values.Add("text", source)
	req, err := http.NewRequest("POST", baseUrl+"/form", strings.NewReader(values.Encode()))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to request to PlantUML server: err=%s", err)
	}
	resp.Body.Close()

	locationUrl, err := resp.Location()
	if err != nil {
		log.Fatal(err)
	}

	umlId := strings.TrimPrefix(locationUrl.Path, "/uml/")
	// log.Println(umlId)

	req, err = http.NewRequest("GET", baseUrl+"/check/"+umlId, nil)
	if err != nil {
		log.Fatal(err)
	}

	resp, err = client.Do(req)
	if err != nil {
		log.Fatalf("Failed to request to PlantUML server: err=%s", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()

	fmt.Println(string(body))
}
