TEST?=$$(go list ./...)
GOFMT_FILES?=$$(find . -name '*.go' | grep -vE './_local')
GO_CMD ?= go
APP_NAME = periphery
BUILD_DIR = $(PWD)/build
SHELL := /bin/bash

all: clean tidy fmt lint security test build

setup:
	@command -v golangci-lint 2>&1 > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@command -v gosec 2>&1 > /dev/null || go install github.com/securego/gosec/v2/cmd/gosec@latest
	@command -v goreleaser 2>&1 > /dev/null || go install github.com/goreleaser/goreleaser@latest

fmt:
	$(GO_CMD)fmt -w $(GOFMT_FILES)

tidy:
	go mod tidy

lint:
	golangci-lint run --timeout 5m

lint-fix:
	golangci-lint run --fix

clean:
	rm -rf ./build ./dist

security:
	gosec -exclude G115 -exclude-dir _local -quiet ./...

test:
	go test -v -timeout 30s -coverprofile=cover.out -cover $(TEST)
	go tool cover -func=cover.out

build:
	goreleaser build --snapshot --clean

build-test:
	echo "Standard binary"
	mkdir -p dist
	CGO_ENABLED=0 GOOS=linux $(GO_CMD) build -ldflags="-s -w" -o dist/periphery_linux_amd64/periphery main.go
	du -hs dist/periphery_linux_amd64/periphery

release:
	goreleaser release --skip=announce,publish,validate --clean

pages:
	@echo "Building documentation with Jekyll..."
	@command -v bundle 2>&1 > /dev/null || (echo "Error: bundler not found. Install with: gem install bundler" && exit 1)
	@if [ ! -f Gemfile ]; then \
		echo "Creating Gemfile..."; \
		cat > Gemfile << 'EOF' ; \
source "https://rubygems.org" ; \
gem "github-pages", group: :jekyll_plugins ; \
gem "jekyll-theme-cayman" ; \
gem "webrick" ; \
EOF \
	fi
	@bundle install --quiet
	@bundle exec jekyll build --source docs --destination _site
	@echo "Documentation built in _site/"
	@echo "To serve locally, run: bundle exec jekyll serve --source docs"

pages-serve:
	@echo "Serving documentation at http://localhost:4000"
	@command -v bundle 2>&1 > /dev/null || (echo "Error: bundler not found. Install with: gem install bundler" && exit 1)
	@if [ ! -f Gemfile ]; then \
		echo "Creating Gemfile..."; \
		cat > Gemfile << 'EOF' ; \
source "https://rubygems.org" ; \
gem "github-pages", group: :jekyll_plugins ; \
gem "jekyll-theme-cayman" ; \
gem "webrick" ; \
EOF \
	fi
	@bundle install --quiet
	bundle exec jekyll serve --source docs --watch --livereload

docker-build:
	docker build -t periphery:latest .

docker-run:
	docker run --rm -v $(PWD)/config.yaml:/etc/periphery/config.yaml periphery:latest

.PHONY: clean test security run fmt tidy lint lint-fix build build-test release pages pages-serve docker-build docker-run
