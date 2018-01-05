PORT=8080
API_PORT=8081
ADMIN_PORT=8082

all:
	test

test:
	echo "TBD"

run:
	GA_TRACKING_ID=$(GA_TRACKING_ID) go run ../util/gen_app_yaml.go --in app.yaml --out app.dist.yaml
	dev_appserver.py --logs_path=/tmp/log.db --storage_path=/tmp/storage.db --search_indexes_path=/tmp/search.db --clear_search_indexes=false --port=$(PORT) --api_port=$(API_PORT) --admin_port=$(ADMIN_PORT) app.dist.yaml

deploy:
	GA_TRACKING_ID=$(GA_TRACKING_ID) go run ../util/gen_app_yaml.go --in app.yaml --out app.dist.yaml
	gcloud --project=$(PROJECT) app deploy app.dist.yaml --version=$(VERSION)
