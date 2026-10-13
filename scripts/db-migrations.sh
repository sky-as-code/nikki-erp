#!/usr/bin/env bash
set -euo pipefail

if [ -z "${1:-}" ] || [ -z "${2:-}" ]; then
	echo "Error: module and cwd parameters are required. Usage: $0 <module> <cwd>" >&2
	exit 1
fi

module="$1"
cwd="$2"

script_path="$(cd "${cwd}scripts" && pwd)"
migration_path_tmp="${script_path}/migrations-tmp"
migration_path="${script_path}/migrations"
map_file="${script_path}/migrations-map.json"

# On Git Bash/MSYS, `pwd` yields "/f/..."; atlas (a native Windows binary) mis-parses that
# inside a file:// URL, dropping the drive letter. cygpath -w gives the real "F:\..." form,
# which atlas wants directly after "file://" (file://F:/..., not file:///F:/...).
if command -v cygpath >/dev/null 2>&1; then
	script_path_url="$(cygpath -w "$script_path" | tr '\\' '/')"
	migration_path_tmp_url="$(cygpath -w "$migration_path_tmp" | tr '\\' '/')"
else
	script_path_url="$script_path"
	migration_path_tmp_url="$migration_path_tmp"
fi

mkdir -p "$migration_path_tmp"
echo "Clearing '${migration_path_tmp}' before generating migration..."
find "$migration_path_tmp" -mindepth 1 -delete

# A cross-module foreign key names a table this module does not own, so atlas must know that
# table or the constraint cannot resolve. The dependency closure's tables go in as a baseline
# migration: atlas replays it, treats those tables as already existing, and diffs only this
# module's tables against the full schema (-withdeps, set in atlas.hcl). Excluding them from the
# diff instead does not work: atlas drops every foreign key that references an excluded table.
baseline_file="${migration_path_tmp}/00000000000000_deps.sql"
(cd "$cwd" && go run -tags staticmods main.go -createsql -dialect=postgres -onlydeps -module="$module") 	> "$baseline_file"
echo "Wrote $(grep -c '^CREATE TABLE' "$baseline_file" || true) dependency table(s) as the diff baseline..."
atlas migrate hash --dir "file://${migration_path_tmp_url}"

# Atlas takes dirs and the config as file:// URLs, the shell needs plain paths for the same locations.
atlas migrate diff tmp 	--dir "file://${migration_path_tmp_url}" 	--config "file://${script_path_url}/atlas.hcl" 	--env nikki 	--var module="$module" 	--var cwd="$cwd"

# Only the generated file; the baseline belongs to the dependencies' own migrations.
sql_files=()
while IFS= read -r sql_file; do
	sql_files+=("$sql_file")
done < <(find "$migration_path_tmp" -maxdepth 1 -type f -name '*_tmp.sql' | sort)

if [ "${#sql_files[@]}" -eq 0 ]; then
	echo "No schema changes detected for module '${module}'; nothing to copy." >&2
	exit 0
fi

target_name=$(grep -o "\"${module}\"[[:space:]]*:[[:space:]]*\"[^\"]*\"" "$map_file" \
	| head -n 1 \
	| sed -e 's/.*:[[:space:]]*"\([^"]*\)"/\1/')

if [ -z "$target_name" ]; then
	echo "Error: module '${module}' is not mapped in ${map_file}." >&2
	exit 1
fi

target_file="${migration_path}/${target_name}.sql"

echo "Copying '${sql_files[0]}' to '${target_file}'..."
cp "${sql_files[0]}" "$target_file"
