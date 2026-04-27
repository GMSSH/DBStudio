.PHONY: all help clean info deps frontend amd64 arm64 package package-amd64 package-arm64

APP_NAME := dbmanager
ENTRY := .
GO := go

CURDIR := $(abspath .)
ROOT_DIR := $(abspath ..)
WEB_DIR := $(ROOT_DIR)/www
WEB_DIST := $(WEB_DIR)/dist

BUILD_DIR := $(CURDIR)/build
CACHE_DIR := /tmp/go-build-cache
TMP_DIR := /tmp/go-tmp

RUNTIME_FILES := config.json
RUNTIME_DIRS := i18n

GOFLAGS_COMMON := -buildvcs=false -trimpath -ldflags="-s -w"

all: package

help:
	@echo ""
	@echo "Available targets:"
	@echo "  make deps          - tidy Go dependencies"
	@echo "  make frontend      - build frontend assets under www/dist"
	@echo "  make amd64         - build linux/amd64 runtime directory"
	@echo "  make arm64         - build linux/arm64 runtime directory"
	@echo "  make package       - build amd64 + arm64 tar.gz packages"
	@echo "  make clean         - remove build outputs"
	@echo "  make info          - show toolchain and output information"
	@echo ""

deps:
	@mkdir -p $(CACHE_DIR) $(TMP_DIR)
	@echo ">>> Tidying Go modules..."
	@env GOCACHE=$(CACHE_DIR) GOTMPDIR=$(TMP_DIR) $(GO) mod tidy

frontend:
	@if [ -d "$(WEB_DIR)" ]; then \
		echo ">>> Building frontend in $(WEB_DIR)..."; \
		cd "$(WEB_DIR)" && pnpm run build; \
		echo ">>> Frontend build complete: $(WEB_DIST)"; \
	else \
		echo ">>> No www directory found, skipping frontend build."; \
	fi

$(BUILD_DIR):
	@mkdir -p $(BUILD_DIR) $(CACHE_DIR) $(TMP_DIR)

define build_arch
	@echo ">>> Building $(1)..."
	@rm -rf $(BUILD_DIR)/$(1)
	@mkdir -p $(BUILD_DIR)/$(1)/app/bin
	@mkdir -p $(BUILD_DIR)/$(1)/app/www
	@mkdir -p $(BUILD_DIR)/$(1)/data
	@mkdir -p $(BUILD_DIR)/$(1)/logs
	@mkdir -p $(BUILD_DIR)/$(1)/tmp
	@env CGO_ENABLED=0 GOOS=linux GOARCH=$(2) GOCACHE=$(CACHE_DIR) GOTMPDIR=$(TMP_DIR) \
		$(GO) build $(GOFLAGS_COMMON) -o $(BUILD_DIR)/$(1)/app/bin/main $(ENTRY)
	@chmod +x $(BUILD_DIR)/$(1)/app/bin/main
	@if command -v upx >/dev/null 2>&1; then \
		echo ">>> Compressing with UPX..."; \
		upx --best --lzma $(BUILD_DIR)/$(1)/app/bin/main || true; \
	else \
		echo ">>> UPX not found, skipping compression."; \
	fi
	@for f in $(RUNTIME_FILES); do \
		if [ -f "$$f" ]; then cp "$$f" $(BUILD_DIR)/$(1)/app/bin/; fi; \
	done
	@for d in $(RUNTIME_DIRS); do \
		if [ -d "$$d" ]; then rm -rf $(BUILD_DIR)/$(1)/app/bin/$$d && cp -r "$$d" $(BUILD_DIR)/$(1)/app/bin/; fi; \
	done
	@if [ -d "$(WEB_DIST)" ]; then \
		rm -rf $(BUILD_DIR)/$(1)/app/www && mkdir -p $(BUILD_DIR)/$(1)/app && cp -r "$(WEB_DIST)" $(BUILD_DIR)/$(1)/app/www; \
	fi
	@echo ">>> Output ready: $(BUILD_DIR)/$(1)"
endef

amd64: $(BUILD_DIR) deps
	$(call build_arch,linux-amd64,amd64)

arm64: $(BUILD_DIR) deps
	$(call build_arch,linux-arm64,arm64)

package-amd64: amd64
	@echo ">>> Packaging linux-amd64..."
	@cd $(BUILD_DIR)/linux-amd64 && COPYFILE_DISABLE=1 tar -czf $(APP_NAME)-linux-amd64.tar.gz app data logs tmp
	@echo ">>> Created $(BUILD_DIR)/linux-amd64/$(APP_NAME)-linux-amd64.tar.gz"

package-arm64: arm64
	@echo ">>> Packaging linux-arm64..."
	@cd $(BUILD_DIR)/linux-arm64 && COPYFILE_DISABLE=1 tar -czf $(APP_NAME)-linux-arm64.tar.gz app data logs tmp
	@echo ">>> Created $(BUILD_DIR)/linux-arm64/$(APP_NAME)-linux-arm64.tar.gz"

package: frontend package-amd64 package-arm64
	@echo ">>> All packages (with frontend) are ready."

clean:
	@echo ">>> Cleaning build outputs..."
	@rm -rf $(BUILD_DIR)

info:
	@echo "Go: $$($(GO) version)"
	@echo "Host arch: $$(uname -m)"
	@echo "Build dir: $(BUILD_DIR)"
	@echo "Frontend dist: $(WEB_DIST)"
