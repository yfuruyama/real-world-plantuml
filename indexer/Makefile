PORT=8083
API_PORT=8084
ADMIN_PORT=8085
RENDERER_SCHEME=http
RENDERER_HOST=localhost
RENDERER_PORT=8086

# Must be set
GITHUB_API_TOKEN=xxx
GCS_BUCKET=xxx

all:
	test

test:
	echo "TBD"

run:
	dev_appserver.py --port=$(PORT) --api_port=$(API_PORT) --admin_port=$(ADMIN_PORT) --logs_path=/tmp/log.db --default_gcs_bucket_name=$(GCS_BUCKET) --env_var GITHUB_API_TOKEN=$(GITHUB_API_TOKEN) --env_var RENDERER_SCHEME=$(RENDERER_SCHEME) --env_var RENDERER_HOST=$(RENDERER_HOST) --env_var RENDERER_PORT=$(RENDERER_PORT) app.yaml

deploy:
	gcloud --project=${PROJECT} app deploy
