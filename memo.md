cat list | perl assemble_ghapi_url.pl | xargs -I{} sh -c 'curl -H "Authorization: token $TOKEN" -s '{}' | jq -c -M .' > contents

# port
web: 8080
indexer: 8081
renderer: 8082
gaelv: 9090

# run
dev_appserver.py app.yaml --port=8080 --logs_path=/tmp/log.db --env_var GITHUB_API_TOKEN=xxx --env_var RENDERER_SCHEME=http --env_var RENDERER_HOST=localhost --env_var RENDERER_PORT=8082 --default_gcs_bucket_name=xxx

# gcs notification
```
gsutil notification create -t projects/${PROJECT_ID}/topics/gcs_notification -f json gs://${BUCKET}
```

```
indexer.PubSubSubscription{Message:indexer.PubSubMessage{Data:"ewogICJraW5kIjogInN0b3JhZ2Ujb2JqZWN0IiwKICAiaWQiOiAicmVhbC13b3JsZC1wbGFudHVtbC91cmxzL3dvcmxkLzE1MTQ1Njc0NTkzMDIxNDIiLAogICJzZWxmTGluayI6ICJodHRwczovL3d3dy5nb29nbGVhcGlzLmNvbS9zdG9yYWdlL3YxL2IvcmVhbC13b3JsZC1wbGFudHVtbC9vL3VybHMlMkZ3b3JsZCIsCiAgIm5hbWUiOiAidXJscy93b3JsZCIsCiAgImJ1Y2tldCI6ICJyZWFsLXdvcmxkLXBsYW50dW1sIiwKICAiZ2VuZXJhdGlvbiI6ICIxNTE0NTY3NDU5MzAyMTQyIiwKICAibWV0YWdlbmVyYXRpb24iOiAiMSIsCiAgImNvbnRlbnRUeXBlIjogImFwcGxpY2F0aW9uL29jdGV0LXN0cmVhbSIsCiAgInRpbWVDcmVhdGVkIjogIjIwMTctMTItMjlUMTc6MTA6NTkuMjk4WiIsCiAgInVwZGF0ZWQiOiAiMjAxNy0xMi0yOVQxNzoxMDo1OS4yOThaIiwKICAic3RvcmFnZUNsYXNzIjogIk1VTFRJX1JFR0lPTkFMIiwKICAidGltZVN0b3JhZ2VDbGFzc1VwZGF0ZWQiOiAiMjAxNy0xMi0yOVQxNzoxMDo1OS4yOThaIiwKICAic2l6ZSI6ICI2IiwKICAibWQ1SGFzaCI6ICJzWlJxeVNTUzBqUjhZalcwMG1FUmhBPT0iLAogICJtZWRpYUxpbmsiOiAiaHR0cHM6Ly93d3cuZ29vZ2xlYXBpcy5jb20vZG93bmxvYWQvc3RvcmFnZS92MS9iL3JlYWwtd29ybGQtcGxhbnR1bWwvby91cmxzJTJGd29ybGQ/Z2VuZXJhdGlvbj0xNTE0NTY3NDU5MzAyMTQyJmFsdD1tZWRpYSIsCiAgImNyYzMyYyI6ICJOVDNZdmc9PSIsCiAgImV0YWciOiAiQ1A3Vm1idmJyOWdDRUFFPSIKfQo=", Attributes:map[string]string{"notificationConfig":"projects/_/buckets/real-world-plantuml/notificationConfigs/1", "bucketId":"real-world-plantuml", "objectId":"urls/world", "objectGeneration":"1514567459302142", "eventType":"OBJECT_FINALIZE", "payloadFormat":"JSON_API_V1"}, MessageId:"18133860677053", PublishTime:"2017-12-29T17:10:59.692Z"}, Subscription:"projects/real-world-plantuml/subscriptions/indexer"}
```
