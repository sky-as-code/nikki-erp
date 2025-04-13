env ?= local
goarch ?= amd64
goos ?= windows
outfile ?= nikki-erp.exe

cwd := $(dir $(lastword $(MAKEFILE_LIST)))

clean:
	@echo "Cleaning dist directory..."
	@rm -rf dist
	@echo "Clean completed"

build-static: clean
	@echo "Building static binary..."
	GOOS=$(goos) GOARCH=$(goarch) go build -work -o dist/$(outfile) cmd/main.go
	@echo "Static build completed"

# Build all modules as plugins
build-mods:
	@echo "Building all modules..."
	@mkdir -p cmd/modules
	@for module_dir in ./modules/*/; do \
		if [ -d "$$module_dir" ]; then \
			module_name=$$(basename "$$module_dir"); \
			echo "Building $$module_name..."; \
			CGO_ENABLED=1 GOOS=$(goos) GOARCH=$(goarch) go build -buildmode=plugin -o ./cmd/modules/$$module_name.so ./modules/$$module_name; \
		fi \
	done
	@echo "All modules built successfully"

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

gen-ent:
	@if [ -z "$(module)" ]; then \
		echo "Error: module parameter is required. Usage: make gen-ent module=<module_name>"; \
		exit 1; \
	fi
	@if [ ! -d "./modules/$(module)/infra/ent" ]; then \
		echo "Error: ent schema directory not found for module '$(module)'"; \
		exit 1; \
	fi
	@echo "Generating ent code for module '$(module)'..."
	go generate ./modules/$(module)/infra/ent

gen-migration:
	@if [ -z "$(module)" ]; then \
		echo "Error: module parameter is required. Usage: make gen-migration module=<module_name>"; \
		exit 1; \
	fi
	@if [ ! -d "./modules/$(module)/infra/ent" ]; then \
		echo "Error: ent schema directory not found for module '$(module)'"; \
		exit 1; \
	fi
	@echo "Generating migration for module '$(module)'..."
	atlas migrate diff migration_name \
		--dir "file://modules/$(module)/migration" \
		--to "ent://modules/$(module)/infra/ent/schema" \
		--dev-url "docker://postgres/17/test?search_path=public"

apply-migration:
	@if [ -z "$(module)" ]; then \
		echo "Error: module parameter is required. Usage: make gen-migration module=<module_name>"; \
		exit 1; \
	fi
	@if [ ! -d "./modules/$(module)/infra/ent" ]; then \
		echo "Error: ent schema directory not found for module '$(module)'"; \
		exit 1; \
	fi
	@echo "Applying migration for module '$(module)'..."
	atlas migrate apply \
		--dir "file://modules/$(module)/migration" \
		--url "postgres://nikki_admin:nikki_password@localhost:5432/nikki_erp?search_path=public&sslmode=disable"

infra-up:
	docker compose -f "${cwd}/scripts/docker/docker-compose.local.yml" up -d

infra-down:
	docker compose -f "${cwd}/scripts/docker/docker-compose.local.yml" down


.PHONY: build build-mods build-static build-dynamic gen-ent gen-migration
