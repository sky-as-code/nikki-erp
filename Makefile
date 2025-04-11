env ?= local
goarch ?= amd64
goos ?= windows
outfile ?= nikki-erp.exe

cwd := $(dir $(lastword $(MAKEFILE_LIST)))

# Build all modules as plugins
build-mods:
	@echo "Building all modules..."
	@mkdir -p cmd/modules
	@while IFS= read -r module_name || [ -n "$$module_name" ]; do \
		echo "Building $$module_name..."; \
		CGO_ENABLED=1 GOOS=$(goos) GOARCH=$(goarch) go build -buildmode=plugin -o ./cmd/modules/$$module_name.so ./modules/$$module_name; \
	done < mods.txt
	@echo "All modules built successfully"

clean:
	@echo "Cleaning dist directory..."
	@rm -rf dist
	@echo "Clean completed"

build-static: clean
	@echo "Building static binary..."
	GOOS=$(goos) GOARCH=$(goarch) go build -work -o dist/$(outfile) cmd/main.go
	@echo "Static build completed"

build-dynamic: build-mods clean
	@echo "Building dynamic binary..."
	GOOS=$(goos) GOARCH=$(goarch) go build -work -tags dynamicmods -o dist/$(outfile) cmd/main.go
	@cp -rf cmd/modules dist/
	@echo "Dynamic build completed"

# Build application and copy config files
build: build-dynamic
	@echo "Copying config files..."
	@mkdir -p dist/config
	@cp cmd/config/config.json dist/config/config.json
	@echo "Build completed successfully"

nikki:
	@[ -f cmd/config/local.env ] || cp cmd/config/local.env.sample cmd/config/local.env
	APP_ENV=$(env) WORKING_DIR="$(cwd)/app" go run cmd/main.go

.PHONY: build build-mods build-static build-dynamic
