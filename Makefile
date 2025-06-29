env ?= local
goarch ?= amd64
goos ?= windows
outfile ?= nikki-erp.exe

cwd := $(dir $(lastword $(MAKEFILE_LIST)))

# START: Go builds

clean:
	@echo "Cleaning dist directory..."
	@rm -rf dist
	@echo "Clean completed"

build-static: clean
	@echo "Building static binary..."
	GOOS=$(goos) GOARCH=$(goarch) go build -tags staticmods -work -o dist/$(outfile) cmd/main.go
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
	GOOS=$(goos) GOARCH=$(goarch) go build -tags dynamicmods -work -tags dynamicmods -o dist/$(outfile) cmd/main.go
	@cp -rf cmd/modules dist/
	@echo "Dynamic build completed"

# Build application and copy config files
build: build-dynamic
	@echo "Copying config files..."
	@mkdir -p dist/config
	@cp cmd/config/config.json dist/config/config.json
	@echo "Build completed successfully"

# END: Go builds

# START: ORM & Database
migration_dir := file://./scripts/migrations

ent-init:
	@if [ -z "$(module)" ]; then \
		echo "Error: module folder is required. Usage: make ent-init module=<module_name> name=<schema_name>"; \
		exit 1; \
	fi
	@if [ -z "$(name)" ]; then \
		echo "Error: Entity name (in PascalCase) is required. Usage: make ent-init module=<module_name> name=<schema_name>"; \
		exit 1; \
	fi
	@module_path="./modules/$(module)/infra/ent/schema"; \
	echo "Initializing ent schema '$(name)' in '$$module_path'..."; \
	go run -mod=mod entgo.io/ent/cmd/ent new $(name) --target $$module_path; \
	printf "package ent\n\n//go:generate go run -mod=mod entgo.io/ent/cmd/ent generate ./schema\n" > "./modules/$(module)/infra/ent/generate.go"

ent-gen:
	@if [ -z "$(module)" ]; then \
		echo "Error: module parameter is required. Usage: make ent-gen module=<module_name>"; \
		exit 1; \
	fi
	@if [ ! -d "./modules/$(module)/infra/ent" ]; then \
		echo "Error: ent schema directory not found for module '$(module)'"; \
		exit 1; \
	fi
	@echo "Generating ent code for module '$(module)'..."
	go generate ./modules/$(module)/infra/ent

ent-current:
	@echo "Generating script of current state to '$(migration_dir)'..."
	atlas migrate diff current_state.tmp \
		--dir "$(migration_dir)" \
		--to "postgres://nikki_admin:nikki_password@localhost:5432/nikki_erp?sslmode=disable" \
		--config file://./scripts/atlas.hcl \
		--env local

ent-hash:
	@echo "Hashing migrations in '$(migration_dir)'..."
	@atlas migrate hash --dir "$(migration_dir)"

ent-migration:
	@if [ -z "$(module)" ]; then \
		echo "Error: module parameter is required. Usage: make ent-migration module=<module_name> name=<name>"; \
		exit 1; \
	fi
	@if [ -z "$(name)" ]; then \
		echo "Error: name parameter is required. Usage: make ent-migration module=<module_name> name=<name>"; \
		exit 1; \
	fi
	@if [ ! -d "./modules/$(module)/infra/ent" ]; then \
		echo "Error: ent schema directory not found for module '$(module)'"; \
		exit 1; \
	fi
	@echo "Generating migration named '$(name)' for module '$(module)' to '$(migration_dir)'..."
	atlas migrate diff $(name) \
		--dir "$(migration_dir)" \
		--to "ent://modules/$(module)/infra/ent/schema" \
		--config file://./scripts/atlas.hcl \
		--env local

ent-apply:
	@echo "Applying migration files in '$(migration_dir)'..."
	atlas migrate apply \
		--dir "$(migration_dir)" \
		--url "postgres://nikki_admin:nikki_password@localhost:5432/nikki_erp?search_path=public&sslmode=disable"

ent-revert:
	@echo "Undoing the LATEST APPLIED migration file in '$(migration_dir)'..."
	atlas migrate down \
		--dir "$(migration_dir)" \
		--url "postgres://nikki_admin:nikki_password@localhost:5432/nikki_erp?search_path=public&sslmode=disable" \
		--config file://./scripts/atlas.hcl \
		--env local

# END: ORM & Database

# START: Local development

infra-up:
	docker compose -f "${cwd}/scripts/docker/docker-compose.local.yml" up -d

infra-down:
	docker compose -f "${cwd}/scripts/docker/docker-compose.local.yml" down -v

install-tools:
	go install go.uber.org/mock/mockgen@latest
# curl -sSf https://atlasgo.sh | sh

nikki:
	@[ -f cmd/config/local.env ] || cp cmd/config/local.env.sample cmd/config/local.env
	APP_ENV=$(env) WORKING_DIR="$(cwd)/cmd" LOG_LEVEL="info" go run -tags=staticmods cmd/*.go

# END: Local development

.PHONY: build build-mods build-static build-dynamic clean ent-init ent-gen ent-current ent-hash ent-migration ent-apply infra-up infra-down nikki
