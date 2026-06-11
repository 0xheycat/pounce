SHELL := /bin/bash

.PHONY: help dev build build-engine build-dashboard test run docker docker-run clean

help:
	@echo "Pounce — make targets:"
	@echo "  make build         # build dashboard, then the engine binary -> engine/pounce"
	@echo "  make test          # engine: go vet + go test ; dashboard: type-check"
	@echo "  make run           # build the engine and serve the dashboard on :7766"
	@echo "  make dev           # run engine + dashboard in watch mode (Go + Node)"
	@echo "  make docker        # build the Docker image (no local Go/Node needed)"
	@echo "  make docker-run    # run the image on http://localhost:7766"
	@echo "  make clean         # remove build artifacts"

build-dashboard:
	cd dashboard && npm install && npm run build

build-engine:
	cd engine && CGO_ENABLED=0 go build -ldflags "-s -w" -o pounce ./cmd/pounce

build: build-dashboard build-engine
	@echo "Built engine/pounce — run it with: cd engine && ./pounce --static ../dashboard/dist"

test:
	cd engine && go vet ./... && go test ./...
	cd dashboard && npm install && npm run lint

run: build-engine
	cd engine && ./pounce --static ../dashboard/dist

dev:
	./scripts/dev.sh

docker:
	docker build -t pounce:local .

docker-run:
	docker run --rm -p 7766:7766 -v pounce-data:/home/pounce/.pounce pounce:local

clean:
	rm -f engine/pounce
	rm -rf dashboard/dist
