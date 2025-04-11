env ?= local
goarch ?= amd64
goos ?= windows
outfile ?= nikki-erp.exe

cwd := $(dir $(lastword $(MAKEFILE_LIST)))

# Build all modules as plugins
build-mods:
	echo "Building all modules..."; \
	mkdir -p cmd/modules
	for dir in ./modules/*/ ; do \
		if [ -f "$$dir/go.mod" ]; then \
			module_name=$$(basename $$dir); \
			echo "Building $$module_name..."; \
			go build -buildmode=plugin -o ./cmd/modules/$$module_name.so $$dir; \
		fi; \
	done; \
	echo "All modules built successfully"

clean:
	rm -rf dist

build-static: clean
	GOOS=$(goos) GOARCH=$(goarch) go build -work -o dist/$(outfile) cmd/main.go

build-dynamic: build-mods clean
	GOOS=$(goos) GOARCH=$(goarch) go build -work -tags dynamicmods -o dist/$(outfile) cmd/main.go
	cp -rf cmd/modules dist/

# Build application and copy config files
build: build-dynamic
	mkdir -p dist/config
	cp cmd/config/config.json dist/config/config.json
	echo "Build completed successfully"

nikki:
	[ -f cmd/config/local.env ] || cp cmd/config/local.env.sample cmd/config/local.env
	APP_ENV=$(env) WORKING_DIR="$(cwd)/app" go run cmd/main.go

.PHONY: build build-mods build-static build-dynamic
