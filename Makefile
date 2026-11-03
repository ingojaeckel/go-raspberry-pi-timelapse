# Entry point for building, testing and running the timelapse app.
#
#   make            vet + test + development build (default)
#   make pi         vet + test + cross-compile for Pi Zero W (ARM v6) into the image module
#   make image      make pi + SD card image (needs Docker)
#   make test       frontend tests + Go vet/tests
#   make run        build the frontend and run the server locally on :8080
#   make loadtest   run the k6 load test against a local build
#
# The Go binary embeds the frontend (src/backend/web), so every Go target
# builds the frontend first. Vite writes straight into src/backend/web/dist.

BACKEND  := src/backend
FRONTEND := src/frontend
# The Pi image ships the app from here (gitignored), see docs/installer.md.
PI_APP   := $(CURDIR)/src/installer/custompios/modules/timelapsepi/filesystem/root/opt/timelapse/timelapse

GIT_COMMIT := $(shell git rev-parse --short HEAD)
BUILT_AT   := $(shell date)
LDFLAGS    := -X 'main.gitCommit=$(GIT_COMMIT)' -X 'main.builtAt=$(BUILT_AT)'

.PHONY: all dev pi image test test-frontend test-go vet frontend run loadtest clean

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

# Release build for the Pi Zero W, into the image module (see image).
pi: test
	cd $(BACKEND) && GOOS=linux GOARCH=arm GOARM=6 CGO_ENABLED=0 \
		go build -v -a -trimpath -o $(PI_APP) -ldflags="-s -w $(LDFLAGS)" .
	file $(PI_APP)
	du $(PI_APP)

# SD card image with the app baked in: src/installer/custompios/workspace/*.zip
image: pi
	DIST_VERSION=$(shell git describe --tags --always --dirty) src/installer/custompios/build-image.sh

run: frontend
	cd $(BACKEND) && go run . -port :8080

loadtest:
	src/perf-testing/k6/run-load-test.sh

clean:
	rm -rf $(BACKEND)/web/dist $(BACKEND)/go-raspberry-pi-timelapse
