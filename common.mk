ifeq ($(V),1)
  Q =
#   unknown flag: --progress
#   PROGRESS = --progress plain
  PROGRESS = 
else
  Q = @
  PROGRESS = 
endif

# Where to push the docker image.
DOCKER_REGISTRY ?= registry.tools.3stripes.net
# This repo's root import path (under GOPATH).
MODULE_ROOT ?= fry.org/cmo/cli

# Which architecture to build - see $(ALL_ARCH) for options.
# if the 'local' rule is being run, detect the ARCH from 'go env'
# if it wasn't specified by the caller.
local: ARCH ?= $(shell go env GOOS)-$(shell go env GOARCH)
clean: ARCH ?= $(shell go env GOOS)-$(shell go env GOARCH)
ARCH ?= linux-amd64

REVISION := $(shell echo $$(git rev-parse --short HEAD) ||echo "Unknown Revision")
BUILDINFO_TAG ?= $(shell echo $$(git describe --long --all | tr '/' '-')$$(git diff-index --quiet HEAD -- || echo '-dirty-'$$(git diff-index -u HEAD | openssl sha1 | cut -c 10-17)))
VCS_TAG := $(shell git describe --tags)
ifeq ($(VCS_TAG),)
VCS_TAG := $(BUILDINFO_TAG)
endif

ifeq ($(VERSION),)
VERSION := $(VCS_TAG)
endif

TAG_LATEST ?= false

platform_temp = $(subst -, ,$(ARCH))
GOOS = $(word 1, $(platform_temp))
GOARCH = $(word 2, $(platform_temp))

# timestamp value that is formatted according to the RFC3339 standard
BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
BUILD_TARGET ?= 

IMAGE_NAME_LC = $(shell echo $(IMAGE_NAME) | tr A-Z a-z)

# Set default base image dynamically for each arch
ifeq ($(GOARCH),amd64)
ifeq ($(DEBUG),true)
	DOCKERFILE ?= ./Dockerfile.$(BIN).debug
	VERSION_SUFFIX := -DBG
	BUILD_CACHE :=
else
	DOCKERFILE ?= ./Dockerfile.$(BIN)
	VERSION_SUFFIX := 
	BUILD_CACHE :=  --no-cache
endif
endif

GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_SHORT_COMMIT=$(shell git rev-parse --short HEAD)

VCS_REF:= $(shell git rev-parse HEAD)
VCS_BRANCH := $(shell git symbolic-ref --short -q HEAD)

.PHONY: all
all: package tag push

PHONY: init 
init: $(TOP_LEVEL)/go.mod ## Initialize the module

$(TOP_LEVEL)/go.mod:
	$(Q)cd $(TOP_LEVEL) && go mod init $(MODULE_ROOT)
	$(Q)cd $(TOP_LEVEL) && go mod tidy

PHONY: build-dirs
build-dirs:
	@mkdir -p $(TOP_LEVEL)/output/$(GOOS)/$(GOARCH)/bin
	@mkdir -p $(TOP_LEVEL)/.go/src/$(PKG) $(TOP_LEVEL)/.go/pkg $(TOP_LEVEL)/.go/bin $(TOP_LEVEL)/.go/std/$(GOOS)/$(GOARCH) $(TOP_LEVEL)/.go/go-build

CLEAN_PHONY_TARGETS = clean-linux-amd64 clean-darwin-amd64 clean-windows-amd64
.PHONY: $(CLEAN_PHONY_TARGETS)
$(CLEAN_PHONY_TARGETS): clean-%:
	$(Q)$(MAKE) --no-print-directory ARCH=$* clean

.PHONY: clean
clean-%:
	$(Q)$(MAKE) --no-print-directory ARCH=$* clean

clean: ## Clean out all generated items
	$(Q)echo "Cleaning binary: ./output/$(GOOS)/$(GOARCH)/bin/${BIN}"
	$(Q)test -e $(TOP_LEVEL)/output/$(GOOS)/$(GOARCH)/bin/${BIN} && rm -r $(TOP_LEVEL)/output/$(GOOS)/$(GOARCH) || true

.PHONY: coverage
coverage: ## Generates the total code coverage of the project
	$(Q)$(eval COVERAGE_DIR=$(shell mktemp -d))
	$(Q)mkdir -p $(COVERAGE_DIR)/tmp
	$(Q)for j in $$(go list ./... | grep -v '/vendor/' | grep -v '/ext/'); do go test -covermode=count -coverprofile=$(COVERAGE_DIR)/$$(basename $$j).out $$j > /dev/null 2>&1; done
	$(Q)echo 'mode: count' > $(COVERAGE_DIR)/tmp/full.out
	$(Q)tail -q -n +2 $(COVERAGE_DIR)/*.out >> $(COVERAGE_DIR)/tmp/full.out
	$(Q)@go tool cover -func=$(COVERAGE_DIR)/tmp/full.out | tail -n 1 | sed -e 's/^.*statements)[[:space:]]*//' -e 's/%//'

.PHONY: package
package: ## Create a docker image of the project
	@echo "Packaging image: $(VERSION) [$(GIT_COMMIT)]"
	$(Q)docker build $(BUILD_CACHE) \
		--label 'org.label-schema.vcs-url=$(VCS_PROTOCOL)://$(VCS_USER):$(VCS_PASSWORD)@$(VCS_URL)' \
		--label 'org.label-schema.vcs-ref=$(VCS_REF)' \
		--label 'org.label-schema.version=$(VERSION)' \
		--label 'org.label-schema.name=$(BIN)', \
		--label 'org.label-schema.build-date=$(BUILD_DATE)' \
		--label 'org.label-schema.vendor=adidas' \
		$(PROGRESS) $(DOCKER_TARGET) \
		-t $(IMAGE_NAME_LC):local -f $(DOCKERFILE) $(TOP_LEVEL)

.PHONY: tag
tag: ## Tag image created by package with latest, git commit and version
	@echo "Tagging image: ${VERSION}${VERSION_SUFFIX} $(GIT_COMMIT)"
	$(Q)docker tag $(IMAGE_NAME_LC):local $(IMAGE_NAME_LC):$(GIT_SHORT_COMMIT)
	$(Q)docker tag $(IMAGE_NAME_LC):local $(IMAGE_NAME_LC):${VERSION}${VERSION_SUFFIX}

.PHONY: push
push: tag ## Push tagged images to docker registry
	@echo "Pushing docker image to registry: ${VERSION}${VERSION_SUFFIX} $(GIT_SHORT_COMMIT)"
#	$(Q)(echo $(BASE64_PASSWORD) | base64 --decode | docker login -u danifv27 --password-stdin $(DOCKER_REGISTRY))
	$(Q)docker push $(IMAGE_NAME_LC):$(GIT_SHORT_COMMIT)
	$(Q)docker push $(IMAGE_NAME_LC):${VERSION}${VERSION_SUFFIX}
#	$(Q)docker logout $(DOCKER_REGISTRY)

.PHONY: test
test: unit_test ## Run all available tests

.PHONY: unit_test
unit_test: ## Run all available unit tests
	go test -v $(shell go list ./... | grep -v /vendor/)

PHONY: labels
labels: ## Show image labels
	$(Q)docker inspect -f '{{ range $$k, $$v := .Config.Labels -}}{{ $$k }}={{printf "%s\n" $$v}}{{ end -}}' $(IMAGE_NAME_LC):local

# GNU make required targets declared as .PHONY to be explicit
BUILD_PHONY_TARGETS = build-linux-amd64 build-darwin-amd64 build-windows-amd64
.PHONY: $(BUILD_PHONY_TARGETS)
$(BUILD_PHONY_TARGETS): build-%:
	$(Q)$(MAKE) --no-print-directory ARCH=$* build

PHONY: build
build: output/$(GOOS)/$(GOARCH)/bin/$(BIN) ## Build application for a specific arch

output/$(GOOS)/$(GOARCH)/bin/$(BIN): build-dirs
	@echo "Building binary: $@ $(VERSION)"
	GOOS=$(GOOS) \
	GOARCH=$(GOARCH) \
	VERSION=$(VERSION) \
	REVISION=$(REVISION) \
	BRANCH=$(VCS_BRANCH) \
	BUILDUSER=$(VCS_USER) \
	PKG=$(PKG) \
	BIN=$(BIN) \
	DEBUG=$(DEBUG) \
	OUTPUT_DIR=$(TOP_LEVEL)/output/$(GOOS)/$(GOARCH)/bin \
	$(TOP_LEVEL)/scripts/build.sh

PHONY: local
local: build-dirs ## Build application for the local arch
	GOOS=$(GOOS) \
	GOARCH=$(GOARCH) \
	VERSION=$(VERSION) \
	REVISION=$(REVISION) \
	BRANCH=$(VCS_BRANCH) \
	BUILDUSER=$(VCS_USER) \
	PKG=$(PKG) \
	BIN=$(BIN) \
	OUTPUT_DIR=$(TOP_LEVEL)/output/$(GOOS)/$(GOARCH)/bin \
	DEBUG=$(DEBUG) \
	$(TOP_LEVEL)/scripts/build.sh

ARTIFACTORY_PHONY_TARGETS = artifactory-linux-amd64 artifactory-darwin-amd64
.PHONY: $(ARTIFACTORY_PHONY_TARGETS)
$(ARTIFACTORY_PHONY_TARGETS): artifactory-%:
	$(Q)$(MAKE) --no-print-directory ARCH=$* artifactory

artifactory: ## Push binary to Artifactory
	$(Q)curl -u $(ARTIFACTORY_CREDENTIALS) -X PUT https://<cloudsmith.io>/artifactory/<path>/$(BIN)/$(VERSION)/$(GOOS)/$(GOARCH)/bin/$(BIN) -T $(TOP_LEVEL)/output/$(GOOS)/$(GOARCH)/bin/$(BIN)

artifactory-docs: ## Push doc to Artifactory
	$(Q)curl -u $(ARTIFACTORY_CREDENTIALS) -X PUT https://<cloudsmith.io>/artifactory/<path>/$(BIN)/README.md -T $(TOP_LEVEL)/README.md
	$(Q)$(foreach file, $(wildcard $(TOP_LEVEL)/docs/*.md), curl -u $(ARTIFACTORY_CREDENTIALS) -X PUT https://<cloudsmith.io>/artifactory/<path>/$(BIN)/$(VERSION)/docs/$(notdir $(file)) $(file);)
