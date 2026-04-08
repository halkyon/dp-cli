.PHONY: clean build

RELEASES_DIR := releases
GIT_SHA := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

clean:
	rm -rf releases

build: clean
	@mkdir -p $(RELEASES_DIR)/$(GIT_SHA)

	@echo "Building linux static..."
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o $(RELEASES_DIR)/$(GIT_SHA)/dp_amd64 ./cmd/dp

	@echo "Building macOS amd64..."
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o $(RELEASES_DIR)/$(GIT_SHA)/dp_darwin_amd64 ./cmd/dp

	@echo "Building macOS arm64..."
	@GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o $(RELEASES_DIR)/$(GIT_SHA)/dp_darwin_arm64 ./cmd/dp

	@echo "Computing checksums..."
	@cd $(RELEASES_DIR)/$(GIT_SHA) && sha256sum dp_* > sha256sums.txt

	@echo "Done. Outputs in $(RELEASES_DIR)/$(GIT_SHA)/"
	@ls -la $(RELEASES_DIR)/$(GIT_SHA)/
