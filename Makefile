MODULE = "github.com/szykes/go-main-for-k8s"
APP_VERSION ?= "v0.0.0"
COMMIT_HASH = $(shell git rev-parse --short HEAD)
BUILD_TIMESTAMP = $(shell date '+%Y-%m-%dT%H:%M:%S')

LDFLAGS = -X $(MODULE)/buildinfo.appVersion=$(APP_VERSION) \
          -X $(MODULE)/buildinfo.commitHash=$(COMMIT_HASH) \
          -X $(MODULE)/buildinfo.buildTimestamp=$(BUILD_TIMESTAMP)

GO_BUILD_ENVS = CGO_ENABLED=0
GO_BUILD_FLAGS = -trimpath -ldflags="$(LDFLAGS)"

GO_RUN_ENVS = DEBUG_MODE=true ENV_VARS_FILE=.env $(GO_BUILD_ENVS)

.PHONY: build-app
build-app:
	$(GO_BUILD_ENVS) go build $(GO_BUILD_FLAGS) -o bin/app ./cmd/app

.PHONY: run-app-debug
run-app-debug:
	$(GO_RUN_ENVS) go run $(GO_BUILD_FLAGS) ./cmd/app
