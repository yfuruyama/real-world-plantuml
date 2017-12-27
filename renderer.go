package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
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

type GitHubContent struct {
	Sha     string `json:sha`
	Content string `json:content`
}

func main() {
	var port int

	flag.IntVar(&port, "port", 8080, "Port for server")
	flag.Parse()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		var c GitHubContent
		if err := json.Unmarshal([]byte(line), &c); err != nil {
			log.Fatal(err)
		}
		// log.Println(c.Content)

		umlData, err := base64.StdEncoding.DecodeString(c.Content)
		if err != nil {
			log.Fatalf("Failed to parse: content=%s, err=%s", c.Content, err)
		}

		uml := string(umlData)

		baseUrl := fmt.Sprintf("http://localhost:%d", port)
		// log.Println(baseUrl)

		values := url.Values{}
		values.Add("text", uml)
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
		log.Println(umlId)

		// access to "GET /check/{id}" and check the word "(Error)", 7 words

		req, err = http.NewRequest("GET", baseUrl+"/svg/"+umlId, nil)
		if err != nil {
			log.Fatal(err)
		}

		resp, err = client.Do(req)
		if err != nil {
			log.Fatalf("Failed to request to PlantUML server: err=%s", err)
		}
		svg, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		resp.Body.Close()

		file, err := os.Create("svg/" + c.Sha + ".svg")
		if err != nil {
			log.Fatal(err)
		}

		_, err = file.Write(svg)
		if err != nil {
			log.Fatal(err)
		}
		file.Close()
	}
}
