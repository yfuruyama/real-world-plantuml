.PHONY: test

PORT=8083
API_PORT=8084
ADMIN_PORT=8085
RENDERER_BASE_URL=http://localhost:8086
SYNTAX_CHECKER_BASE_URL=http://localhost:8087

# Must be set
GITHUB_API_TOKEN=xxx
GCS_BUCKET=xxx

all:
	test

test:
	go test -v ./...

run:
	GITHUB_API_TOKEN=$(GITHUB_API_TOKEN) RENDERER_BASE_URL=$(RENDERER_BASE_URL) SYNTAX_CHECKER_BASE_URL=$(SYNTAX_CHECKER_BASE_URL) go run ../util/gen_app_yaml.go --in app.yaml --out app.dist.yaml 
	dev_appserver.py --port=$(PORT) --api_port=$(API_PORT) --admin_port=$(ADMIN_PORT) --logs_path=/tmp/log_indexer.db --storage_path=/tmp/storage.db --search_indexes_path=/tmp/search.db --clear_search_indexes=false --default_gcs_bucket_name=$(GCS_BUCKET) app.dist.yaml

notify:
	curl -X POST http://localhost:$(PORT)/_ah/push-handlers/gcs_notification --data '{"message": {"attributes":{"objectId":"$(OBJECT_ID)", "eventType":"OBJECT_FINALIZE"}, "messageId":"xxx"}, "subscription" :"xxx"}'

deploy:
	GITHUB_API_TOKEN=$(GITHUB_API_TOKEN) RENDERER_BASE_URL=$(RENDERER_BASE_URL) SYNTAX_CHECKER_BASE_URL=$(SYNTAX_CHECKER_BASE_URL) go run ../util/gen_app_yaml.go --in app.yaml --out app.dist.yaml 
	gcloud --project=$(PROJECT) app deploy app.dist.yaml --version=$(VERSION)

deploy_queue:
	gcloud --project=$(PROJECT) app deploy queue.yaml
