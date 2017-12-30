PORT=8080
API_PORT=8081
ADMIN_PORT=8082

all:
	test

test:
	echo "TBD"

run:
	dev_appserver.py --port=$(PORT) --api_port=$(API_PORT) --admin_port=$(ADMIN_PORT) app.yaml

deploy:
	gcloud --project=${PROJECT} app deploy
