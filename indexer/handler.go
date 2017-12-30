package indexer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/appengine"
	"google.golang.org/appengine/file"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
)

type PubSubSubscription struct {
	Message      PubSubMessage `json:message`
	Subscription string        `json:subscription`
}

type PubSubMessage struct {
	Data        string            `json:data`
	Attributes  map[string]string `json:attributes`
	MessageId   string            `json:messageId`
	PublishTime string            `json:publishTime`
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
			Path:    "/indexer/create",
			Payload: bodyBytes,
			Header:  header,
			Method:  "POST",
			Delay:   time.Second * time.Duration(i),
		}
		taskqueue.Add(ctx, task, "index-create-queue")

		i++
	}

	fmt.Fprintf(w, "ok")
}
