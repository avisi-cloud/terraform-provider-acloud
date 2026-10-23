# Contributor workflow for the avisi-cloud/acloud Terraform provider.
#
#   make            show this help
#   make check      everything CI enforces on a pull request
#   make docs       regenerate docs/ from the schema, templates/ and examples/
#   make dev        build the provider and print how to use the local build
#
# Releases are built by goreleaser (.goreleaser.yaml) when a tag is pushed, not by this Makefile.

GO           ?= go
TERRAFORM    ?= terraform
BINARY       := terraform-provider-acloud
BIN_DIR      := $(CURDIR)/bin
TF_RC        := $(BIN_DIR)/dev.terraformrc
TFPLUGINDOCS := $(GO) run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@v0.21.0
EXAMPLES     := $(sort $(dir $(wildcard examples/*/main.tf examples/provider/provider.tf)))

# Used by `make install`, the older way to run a local build: a plugin directory instead of dev_overrides.
HOSTNAME  ?= avisi.cloud
NAMESPACE ?= ame
NAME      ?= acloud
VERSION   ?= 0.2
OS_ARCH   ?= $(shell $(GO) env GOOS)_$(shell $(GO) env GOARCH)

.DEFAULT_GOAL := help
.PHONY: help build dev install test testacc vet fmt fmt-check docs docs-check examples-validate check clean

help: ## Show this help
	@grep -hE '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

## --- build -----------------------------------------------------------------

build: ## Build the provider into bin/
	$(GO) build -o $(BIN_DIR)/$(BINARY) .

dev: build ## Build and write bin/dev.terraformrc, so terraform uses the local build
	@printf 'provider_installation {\n  dev_overrides {\n    "avisi-cloud/acloud" = "%s"\n  }\n  direct {}\n}\n' "$(BIN_DIR)" > $(TF_RC)
	@echo "Use the local build with:  export TF_CLI_CONFIG_FILE=$(TF_RC)"
	@echo "terraform plan/apply then use bin/$(BINARY); no terraform init needed."

install: ## Install into ~/.terraform.d/plugins (override HOSTNAME, NAMESPACE, VERSION for the address)
	$(GO) build -o $(BINARY) .
	mkdir -p ~/.terraform.d/plugins/$(HOSTNAME)/$(NAMESPACE)/$(NAME)/$(VERSION)/$(OS_ARCH)
	mv $(BINARY) ~/.terraform.d/plugins/$(HOSTNAME)/$(NAMESPACE)/$(NAME)/$(VERSION)/$(OS_ARCH)

## --- test ------------------------------------------------------------------

test: ## Run the unit tests
	$(GO) test ./... $(TESTARGS) -timeout=30s -parallel=4

testacc: ## Run the acceptance tests against a real API (creates real resources)
	TF_ACC=1 $(GO) test ./... -v $(TESTARGS) -timeout 120m

vet: ## Run go vet
	$(GO) vet ./...

## --- formatting ------------------------------------------------------------

fmt: ## Format the Go code and every .tf file under examples/
	gofmt -w .
	$(TERRAFORM) fmt -recursive examples

fmt-check: ## Fail if any Go or .tf file is not formatted
	@test -z "$$(gofmt -l .)" || { echo "gofmt needed for:"; gofmt -l .; exit 1; }
	$(TERRAFORM) fmt -check -recursive examples

## --- documentation ---------------------------------------------------------

docs: ## Regenerate docs/ with tfplugindocs
	$(TFPLUGINDOCS) generate

docs-check: docs ## Fail if docs/ is out of date or invalid (use in CI)
	$(TFPLUGINDOCS) validate
	@git diff --exit-code -- docs || { echo "docs/ is out of date: run 'make docs' and commit the result"; exit 1; }
	@test -z "$$(git ls-files --others --exclude-standard docs)" || { echo "new files in docs/: run 'make docs' and commit them"; exit 1; }

examples-validate: dev ## Run terraform validate on every example against the local build
	@for dir in $(EXAMPLES); do \
		echo "==> $$dir"; \
		TF_CLI_CONFIG_FILE=$(TF_RC) $(TERRAFORM) -chdir=$$dir validate -no-color || exit 1; \
	done

check: fmt-check vet test docs-check ## Everything CI enforces on a pull request

clean: ## Remove bin/
	rm -rf $(BIN_DIR)
