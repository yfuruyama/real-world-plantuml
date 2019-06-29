PORT=8080
API_PORT=8081
ADMIN_PORT=8082
DATASTORE_PORT=8083

all:
	test

test:
	echo "TBD"

run-datastore:
	gcloud beta emulators datastore start --host-port=localhost:$(DATASTORE_PORT)

run:
	dev_appserver.py --datastore_emulator_port=$(DATASTORE_PORT) --logs_path=/tmp/log_web.db --search_indexes_path=/tmp/search.db --clear_search_indexes=false --port=$(PORT) --api_port=$(API_PORT) --admin_port=$(ADMIN_PORT) app.yaml

deploy:
	gcloud --project=$(PROJECT) app deploy app.yaml --version=$(VERSION)
