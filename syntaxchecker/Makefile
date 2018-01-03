PORT=8087

.PHONY: build
all: build

build:
	docker build -t syntaxchecker .

run:
	docker run -p $(PORT):8080 syntaxchecker

deploy:
	gcloud --project=$(PROJECT) app deploy
