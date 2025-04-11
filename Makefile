env ?= local
goarch ?= amd64
goos ?= windows
outfile ?= nikki-erp.exe

cwd := $(dir $(lastword $(MAKEFILE_LIST)))

# Build all modules as plugins
build-mods:
	echo "Building all modules..."; \
	mkdir -p main/modules
	for dir in ./modules/*/ ; do \
		if [ -f "$$dir/go.mod" ]; then \
			module_name=$$(basename $$dir); \
			echo "Building $$module_name..."; \
			go build -buildmode=plugin -o ./main/modules/$$module_name.so $$dir; \
		fi; \
	done; \
	echo "All modules built successfully"

clean:
	rm -rf dist

build-static: clean
	GOOS=$(goos) GOARCH=$(goarch) go build -work -o dist/$(outfile) main/main.go

build-dynamic: build-mods clean
	GOOS=$(goos) GOARCH=$(goarch) go build -work -tags dynamicmods -o dist/$(outfile) main/main.go
	cp -rf main/modules dist/

# Build application and copy config files
build: build-dynamic
	mkdir -p dist/config
	cp main/config/config.json dist/config/config.json
	echo "Build completed successfully"

nikki:
	[ -f main/config/local.env ] || cp main/config/local.env.sample main/config/local.env
	APP_ENV=$(env) WORKING_DIR="$(cwd)/app" go run main/main.go

.PHONY: build build-mods build-static build-dynamic
