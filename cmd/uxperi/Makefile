ifeq ($(V),1)
  Q =
else
  Q = @
endif

# The binary to build (just the basename).
BIN ?= uxperi
BASE64_PASSWORD = ZHVtbXlWYWx1ZQ==
IMAGE_NAME = $(DOCKER_REGISTRY)/ebtpc-commons/$(BIN)
VCS_USER ?= cmo-sre
VCS_PASSWORD = $(shell echo $(BASE64_PASSWORD) | base64 --decode)
VCS_PROTOCOL ?= https
VCS_URL ?= https://github.com/danifv27/synthetos
ARTIFACTORY_CREDENTIALS ?= $(shell echo cGMtbWF2ZW46RGVjZW1iZXJAMjAxNg==| base64 --decode)

TOP_LEVEL = ../..

include $(TOP_LEVEL)/common.mk

.PHONY: help
help: ## Show This Help
	@for line in $$(cat ./Makefile $(TOP_LEVEL)/common.mk | grep "##" | grep -v "grep" | sed  "s/:.*##/:/g" | sed "s/\ /!/g"); do verb=$$(echo $$line | cut -d ":" -f 1); desc=$$(echo $$line | cut -d ":" -f 2 | sed "s/!/\ /g"); printf "%-30s--%s\n" "$$verb" "$$desc"; done
