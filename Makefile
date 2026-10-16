HOSTNAME=avisi.cloud
NAMESPACE=ame
NAME=acloud
BINARY=terraform-provider-${NAME}
VERSION=0.2
OS_ARCH?=$(shell go env GOOS)_$(shell go env GOARCH)

# Pinned so CI results only change when this version (or the vulnerability database) changes.
# renovate: datasource=go depName=golang.org/x/vuln
GOVULNCHECK_VERSION=v1.9.0

.PHONY: default build release install test testrace testacc vet fmt fmtcheck tidycheck vulncheck check

default: install

build:
	go build -o ${BINARY}

release:
	GOOS=darwin GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_darwin_amd64
	GOOS=freebsd GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_freebsd_386
	GOOS=freebsd GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_freebsd_amd64
	GOOS=freebsd GOARCH=arm go build -o ./bin/${BINARY}_${VERSION}_freebsd_arm
	GOOS=linux GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_linux_386
	GOOS=linux GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_linux_amd64
	GOOS=linux GOARCH=arm go build -o ./bin/${BINARY}_${VERSION}_linux_arm
	GOOS=openbsd GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_openbsd_386
	GOOS=openbsd GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_openbsd_amd64
	GOOS=solaris GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_solaris_amd64
	GOOS=windows GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_windows_386
	GOOS=windows GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_windows_amd64

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

# Unit tests only: TF_ACC is cleared so acceptance tests are skipped even if it is set in your shell.
test:
	TF_ACC= go test ./... $(TESTARGS) -timeout=5m

testrace:
	TF_ACC= go test -race ./... $(TESTARGS) -timeout=10m

# Acceptance tests create and destroy real Avisi Cloud resources and require credentials
# (ACLOUD_PERSONAL_ACCESS_TOKEN, ACLOUD_ORGANISATION, optionally ACLOUD_API_ENDPOINT).
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

vet:
	go vet ./...

fmt:
	gofmt -w .

fmtcheck:
	@unformatted="$$(gofmt -l .)" || exit $$?; \
	if [ -n "$$unformatted" ]; then \
		echo "The following files are not gofmt-formatted (run 'make fmt'):"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

tidycheck:
	go mod tidy -diff

vulncheck:
	go run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) ./...

check: fmtcheck tidycheck vet test testrace vulncheck
