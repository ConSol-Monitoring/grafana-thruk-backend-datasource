PLUGINNAME=consolmonitoring-thruk-datasource
TAGVERSION=$(shell git describe --tag --exact-match 2>/dev/null | sed -e 's/^v//')
DOCKER=docker run \
		-t \
		--rm \
		-v $(shell pwd):/src \
		-w "/src" \
		-u $(shell id -u):$(shell id -g) \
		-e "HOME=/src" \
		-e "GRAFANA_ACCESS_POLICY_TOKEN=$(GRAFANA_ACCESS_POLICY_TOKEN)"
NODEVERSION=24
export NODE_PATH=$(shell pwd)/node_modules
SHELL=bash

# golangci-lint is used for linting the backend
# govulncheck is used for scanning the backend for known vulnerabilities
# mage has to be installed
GO          ?= go
GOBIN       ?= $(shell $(GO) env GOPATH)/bin
GOLANGCI    ?= $(GOBIN)/golangci-lint
GOVULNCHECK ?= $(GOBIN)/govulncheck
# govulncheck v1.1.4 cannot analyze the go1.27 standard library. Pin the
# toolchain to the same Go 1.26 line used by the plugin's build and the
# official plugin validator.
GOTOOLCHAINVERSION ?= go1.26.6
PROJECT     ?= grafana-thruk-backend-datasource

# current go version as major*10000+minor, ex.: go1.27.0 -> 10027
GOVERSION        := $(shell $(GO) version | awk '/^go version go/{ v = $$3; sub(/^go/, "", v); split(v, a, "."); print a[1]*10000 + a[2] }')
MINGOVERSION     := 10026
MINGOVERSIONSTR  := 1.26

build:
	$(DOCKER)    --name $(PLUGINNAME)-build        node:$(NODEVERSION) bash -c "npm install && npm run build"

buildwatch:
	$(DOCKER) -i --name $(PLUGINNAME)-buildwatch   node:$(NODEVERSION) bash -c "npm install && npm run dev"

buildupgrade:
	rm -f package-lock.json
	$(DOCKER)    --name $(PLUGINNAME)-buildupgrade node:$(NODEVERSION) bash -c "npm install && npm update $(filter-out $@,$(MAKECMDGOALS))"

buildaudit:
	$(DOCKER)    --name $(PLUGINNAME)-buildaudit   node:$(NODEVERSION) bash -c "npm install && npm audit"

buildsign:
	$(DOCKER)    --name $(PLUGINNAME)-buildsign    node:$(NODEVERSION) bash -c "npm install && npx @grafana/sign-plugin"

buildnpm:
	$(DOCKER)    --name $(PLUGINNAME)-buildnpm     node:$(NODEVERSION) bash -c "npm $(filter-out $@,$(MAKECMDGOALS))"

prettier:
	$(DOCKER)    --name $(PLUGINNAME)-buildpret    node:$(NODEVERSION) npx prettier --write --ignore-unknown src/

prettiercheck:
	$(DOCKER)    --name $(PLUGINNAME)-buildprtchck node:$(NODEVERSION) npx prettier --check --ignore-unknown src/

buildshell:
	$(DOCKER) -i --name $(PLUGINNAME)-buildshell   node:$(NODEVERSION) bash

jesttest:
	$(DOCKER)    --name $(PLUGINNAME)-buildprtchck node:$(NODEVERSION) npx jest test

test: build prettiercheck jesttest

# start a specific grafana version like:
# GRAFANA_VERSION=12.4.0 make dev
dev:
	@mkdir -p dist
	docker compose build
	docker compose up

clean:
	-docker compose rm -f
	-sudo chown $(shell id -u):$(shell id -g) -R dist node_modules
	rm -rf dist
	rm -rf node_modules
	rm -rf .yarnrc
	rm -rf .npm

releasebuild:
	@if [ "x$(TAGVERSION)" = "x" ]; then echo "ERROR: must be on a git tag, got: $(shell git describe --tag --dirty)"; exit 1; fi
	$(MAKE) clean
	$(MAKE) build
	$(MAKE) GRAFANA_ACCESS_POLICY_TOKEN=$(GRAFANA_ACCESS_POLICY_TOKEN) buildsign
	mv dist/ $(PLUGINNAME)
	rm -f $(PLUGINNAME)-$(TAGVERSION).zip
	zip $(PLUGINNAME)-$(TAGVERSION).zip $(PLUGINNAME) -r
	rm -rf $(PLUGINNAME)
	@echo "release build successful: $(TAGVERSION)"
	ls -la $(PLUGINNAME)-$(TAGVERSION).zip

buildbackend:
	$(GOBIN)/mage -v
	chmod 0755 dist/gpx_*

# just skip unknown make targets
.DEFAULT:
	@if [[ "$(MAKECMDGOALS)" =~ ^buildupgrade ]] || [[  "$(MAKECMDGOALS)" =~ ^buildnpm ]] ; then \
		: ; \
	else \
		echo "unknown make target(s): $(MAKECMDGOALS)"; \
		exit 1; \
	fi

versioncheck:
	@if [ -z "$(GOVERSION)" ] || [ "$(GOVERSION)" -lt "$(MINGOVERSION)" ]; then \
		echo "**** ERROR:"; \
		echo "**** $(PROJECT) requires at least golang version $(MINGOVERSIONSTR) or higher"; \
		echo "**** this is: $$($(GO) version)"; \
		exit 1; \
	fi

tools: | versioncheck
	@if [ ! -x "$(GOLANGCI)" ]; then \
		echo "installing golangci-lint ..."; \
		( cd buildtools && $(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint ); \
	fi
	@if [ ! -x "$(GOVULNCHECK)" ]; then \
		echo "installing govulncheck ..."; \
		( cd buildtools && $(GO) install golang.org/x/vuln/cmd/govulncheck ); \
	fi

govulncheck: tools
	GOTOOLCHAIN=$(GOTOOLCHAINVERSION) $(GOVULNCHECK) ./...

golangci: tools
	$(GOLANGCI) version
	@echo "  - GOOS=linux"
	GOOS=linux CGO_ENABLED=0 $(GOLANGCI) run $(GOLANG_CI_OPTIONS) ./...

# Package the built plugin as plugin.zip
# This is intdended to be run with grafana plugin-validator
build-plugin-for-validation: build
	@$(MAKE) buildbackend
	@set -eu; \
		PLUGIN_ID=$$(grep '"id"' < src/plugin.json | sed -E 's/.*"id" *: *"(.*)".*/\1/' | tr -cd 'a-zA-Z0-9._-'); \
		rm -rf "$${PLUGIN_ID}" plugin.zip; \
		cp -r dist "$${PLUGIN_ID}"; \
		zip -qr plugin.zip "$${PLUGIN_ID}"; \
		rm -rf "$${PLUGIN_ID}"; \
		echo "created plugin.zip"

# Package the plugin source as sources.zip (overwrites any existing sources.zip).
# Excludes caches, build output, git metadata and other files that are not part of the plugin source.
# This is intdended to be run with grafana plugin-validator
build-plugin-sources-for-validation:
	rm -f sources.zip
	cd .. && zip -r "$(notdir $(CURDIR))/sources.zip" "$(notdir $(CURDIR))" \
		-x '$(notdir $(CURDIR))/.cache/*' \
		-x '$(notdir $(CURDIR))/.npm/*' \
		-x '$(notdir $(CURDIR))/node_modules/*' \
		-x '$(notdir $(CURDIR))/.git/*' \
		-x '$(notdir $(CURDIR))/dist/*' \
		-x '$(notdir $(CURDIR))/logs/*' \
		-x '$(notdir $(CURDIR))/test-results/*' \
		-x '$(notdir $(CURDIR))/playwright-report/*' \
		-x '$(notdir $(CURDIR))/.eslintcache' \
		-x '$(notdir $(CURDIR))/pre-release-check.md' \
		-x '$(notdir $(CURDIR))/.bra.toml' \
		-x '$(notdir $(CURDIR))/TODO' \
		-x '$(notdir $(CURDIR))/*.zip'
	@echo "created sources.zip"
