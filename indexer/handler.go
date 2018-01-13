package indexer

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/appengine"
	"google.golang.org/appengine/file"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
	"google.golang.org/appengine/urlfetch"
)

type IndexCreateRequestBody struct {
	Url string `json:"url"`
}

type GitHubContentResponse struct {
	Path    string `json:"path"`
	Sha     string `json:"sha"`
	Content string `json:"content"`
}

type PubSubSubscription struct {
	Message      PubSubMessage `json:message`
	Subscription string        `json:subscription`
}

type PubSubMessage struct {
	Attributes map[string]string `json:attributes`
	MessageId  string            `json:messageId`
}

func HandleIndexCreate(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	decoder := json.NewDecoder(r.Body)
	var body IndexCreateRequestBody
	if err := decoder.Decode(&body); err != nil {
		log.Warningf(ctx, "%s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Infof(ctx, "url: %s", body.Url)

	re := regexp.MustCompile(`^https://github.com/([^/]+)/([^/]+)/blob/([^/]+)/(.+)$`)
	matched := re.FindStringSubmatch(body.Url)
	if len(matched) != 5 {
		log.Warningf(ctx, "invalid github url")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	owner := matched[1]
	repo := matched[2]
	hash := matched[3]
	path := matched[4]

	apiUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s", owner, repo, path, hash)

	token := os.Getenv("GITHUB_API_TOKEN")
	req, _ := http.NewRequest("GET", apiUrl, nil)
	req.Header.Add("Authorization", fmt.Sprintf("token %s", token))

	client := urlfetch.Client(ctx)
	resp, err := client.Do(req)
	if err != nil {
		log.Criticalf(ctx, "Failed to request to GitHub: err=%s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	decoder = json.NewDecoder(resp.Body)
	var ghcResp GitHubContentResponse
	if err := decoder.Decode(&ghcResp); err != nil {
		log.Criticalf(ctx, "Failed to parse response: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Infof(ctx, "Get content response: %#v", ghcResp)
	contentBytes, err := base64.StdEncoding.DecodeString(ghcResp.Content)
	if err != nil {
		log.Criticalf(ctx, "Failed to parse GitHub content: err=%s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	content := string(contentBytes)

	rendererBaseUrl := os.Getenv("RENDERER_BASE_URL")
	renderer := NewRenderer(ctx, rendererBaseUrl)

	syntaxCheckerBaseUrl := os.Getenv("SYNTAX_CHECKER_BASE_URL")
	syntaxChecker := NewSyntaxChecker(ctx, syntaxCheckerBaseUrl)

	indexer := NewIndexer(renderer, syntaxChecker)
	err = indexer.CreateIndexes(ctx, content, body.Url)
	if err != nil {
		log.Criticalf(ctx, "%s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "ok")
}

func HandleGcsNotification(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	decoder := json.NewDecoder(r.Body)
	var sub PubSubSubscription
	if err := decoder.Decode(&sub); err != nil {
		log.Warningf(ctx, "%s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Infof(ctx, "Received: %#v", sub)

	// TODO: more sophisticated guard
	if _, ok := sub.Message.Attributes["eventType"]; !ok {
		log.Warningf(ctx, "Not GCS notification")
		w.WriteHeader(http.StatusOK)
		return
	}

	typ := sub.Message.Attributes["eventType"]
	objectId := sub.Message.Attributes["objectId"]

	if typ != "OBJECT_FINALIZE" {
		log.Infof(ctx, "Not OBJECT_FINALIZE event: %s", typ)
		w.WriteHeader(http.StatusOK)
		return
	}

	bucketName, err := file.DefaultBucketName(ctx)
	if err != nil {
		log.Criticalf(ctx, "failed to get default GCS bucket name: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Criticalf(ctx, "failed to init gcs client: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer client.Close()

	reader, err := client.Bucket(bucketName).Object(objectId).NewReader(ctx)
	if err != nil {
		log.Criticalf(ctx, "unable to open file from bucket %v, object %v: %v", bucketName, objectId, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	i := 0
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		log.Infof(ctx, "Read line: %s", line)

		header := make(http.Header)
		header.Set("Content-Type", "application/json")

		body := &IndexCreateRequestBody{
			Url: line,
		}
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			log.Criticalf(ctx, "json marshal error: %s", err)
			continue
		}

		task := &taskqueue.Task{
			Path:    "/indexes",
			Payload: bodyBytes,
			Header:  header,
			Method:  "POST",
			Delay:   5 * time.Second * time.Duration(i),
		}
		taskqueue.Add(ctx, task, "index-create-queue")

		i++
	}

	fmt.Fprintf(w, "ok")
}
