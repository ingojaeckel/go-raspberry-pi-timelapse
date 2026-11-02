# Entry point for building, testing and running the timelapse app.
#
#   make            vet + test + development build (default)
#   make pi         vet + test + cross-compile for Pi Zero W (ARM v6)
#   make test       frontend tests + Go vet/tests
#   make run        build the frontend and run the server locally on :8080
#   make loadtest   run the k6 load test against a local build
#
# The Go binary embeds the frontend (src/backend/web), so every Go target
# builds the frontend first. Vite writes straight into src/backend/web/dist.

BACKEND  := src/backend
FRONTEND := src/frontend
PI_BIN   := bin/arm/go-raspberry-pi-timelapse

GIT_COMMIT := $(shell git rev-parse --short HEAD)
BUILT_AT   := $(shell date)
LDFLAGS    := -X 'main.gitCommit=$(GIT_COMMIT)' -X 'main.builtAt=$(BUILT_AT)'

.PHONY: all dev pi test test-frontend test-go vet frontend run loadtest clean

all: dev

## Frontend ------------------------------------------------------------------

$(FRONTEND)/node_modules: $(FRONTEND)/package.json $(FRONTEND)/package-lock.json
	cd $(FRONTEND) && npm ci
	@touch $@

frontend: $(FRONTEND)/node_modules
	cd $(FRONTEND) && ./build.sh

test-frontend: $(FRONTEND)/node_modules
	cd $(FRONTEND) && npm test

## Backend -------------------------------------------------------------------

vet: frontend
	cd $(BACKEND) && go vet ./...

# -race needs cgo, which the Pi cross-compile does not use; tests run natively.
test-go: frontend
	cd $(BACKEND) && go test -race -shuffle=on -cover ./...

test: test-frontend vet test-go

# Development build for quick local iteration.
dev: test
	cd $(BACKEND) && go build -v -trimpath -o go-raspberry-pi-timelapse -ldflags="$(LDFLAGS)" .

# Release build for the Pi. The committed binary at src/backend/bin/arm is what
# the Pi image runs (see src/installer/custompios/.../timelapse.service).
pi: test
	cd $(BACKEND) && GOOS=linux GOARCH=arm GOARM=6 CGO_ENABLED=0 \
		go build -v -a -trimpath -o $(PI_BIN) -ldflags="-s -w $(LDFLAGS)" .
	file $(BACKEND)/$(PI_BIN)
	du $(BACKEND)/$(PI_BIN)

run: frontend
	cd $(BACKEND) && go run . -port :8080

loadtest:
	src/perf-testing/k6/run-load-test.sh

clean:
	rm -rf $(BACKEND)/web/dist $(BACKEND)/go-raspberry-pi-timelapse
