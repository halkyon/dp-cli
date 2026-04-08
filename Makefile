.PHONY: clean build

RELEASES_DIR := releases
GIT_SHA := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "")

ifeq ($(GIT_TAG),)
	VERSION := $(GIT_SHA)
else
	VERSION := $(GIT_TAG)
endif

LDFLAGS := -s -w -X=main.version=$(VERSION)

clean:
	rm -rf releases

build: clean
	@mkdir -p $(RELEASES_DIR)/$(GIT_SHA)

	@echo "Building linux static..."
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(RELEASES_DIR)/$(GIT_SHA)/dp_amd64 ./cmd/dp

	@echo "Building macOS amd64..."
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(RELEASES_DIR)/$(GIT_SHA)/dp_darwin_amd64 ./cmd/dp

	@echo "Building macOS arm64..."
	@GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(RELEASES_DIR)/$(GIT_SHA)/dp_darwin_arm64 ./cmd/dp

	@echo "Computing checksums..."
	@cd $(RELEASES_DIR)/$(GIT_SHA) && sha256sum dp_* > sha256sums.txt

	@echo "Done. Outputs in $(RELEASES_DIR)/$(GIT_SHA)/"
