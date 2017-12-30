PORT=8086
CONTAINER_PORT=8080

all: build

build:
	docker build -t plantuml-renderer .

run:
	docker run -p $(PORT):$(CONTAINER_PORT) plantuml-renderer
