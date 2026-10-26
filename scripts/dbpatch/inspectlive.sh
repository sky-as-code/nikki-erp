#!/usr/bin/env bash
# Stage 1 of 3: snapshot the public schema of a running database, read-only (catalog queries only, no rows,
# no locks). Writes live_schema.sql, inspect.env and REPORT.md to <cwd>scripts/dbpatch-out/<ts>/.
# Usage: inspectlive.sh <cwd> [url]    (url defaults to $DBPATCH_URL)
set -euo pipefail

source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

# Atlas keeps its revisions in "public" when the URL sets search_path=public (as `make db-apply` does),
# otherwise in its own "atlas_schema_revisions" schema.
find_revisions_schema() {
	dbp_psql "$dbp_url" -At -c "
		SELECT n.nspname FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE c.relname = 'atlas_schema_revisions' AND c.relkind = 'r'
		ORDER BY n.nspname = 'public' DESC LIMIT 1"
}

write_inspect_env() {
	local run_dir="$1" schema="$2"
	echo "inspected_at=$(date '+%Y-%m-%d %H:%M:%S %z')"
	echo "atlas_version=$(atlas version 2>/dev/null | head -n 1)"
	echo "host=${dbp_url_host}"
	echo "port=${dbp_url_port}"
	echo "database=${dbp_url_db}"
	echo "server_version=$(dbp_psql "$dbp_url" -At -c 'SHOW server_version')"
	echo "tree=${dbp_migration_dir}"
	echo "revisions_schema=${schema}"
	echo "live_schema_sha256=$(dbp_sha "${run_dir}/live_schema.sql")"
}

write_report() {
	local env_file="$1/inspect.env"
	echo "# DB patch report"
	echo
	echo "- Live: \`${dbp_url_label}\`, PostgreSQL $(dbp_env_get "$env_file" server_version)"
	echo "- Tree: \`${dbp_migration_dir}\`"
	echo "- Inspected: $(dbp_env_get "$env_file" inspected_at)"
}

main() {
	[ "$#" -ge 1 ] || dbp_die "usage: $0 <cwd> [url]"
	dbp_resolve_tree "$1"
	dbp_resolve_url "${2:-}"
	dbp_require_tools atlas docker sha256sum
	local run_dir schema
	run_dir="${dbp_out_root}/$(date +%Y%m%d-%H%M%S)"
	mkdir -p "$run_dir"

	dbp_log "Inspecting the public schema of ${dbp_url_label} (read-only)..."
	dbp_atlas_inspect "$dbp_url" > "${run_dir}/live_schema.sql"
	[ -s "${run_dir}/live_schema.sql" ] || dbp_die "atlas returned an empty schema for ${dbp_url_label}."
	schema="$(find_revisions_schema)"
	write_inspect_env "$run_dir" "$schema" > "${run_dir}/inspect.env"
	write_report "$run_dir" > "${run_dir}/REPORT.md"

	dbp_log "Inspection written to ${run_dir}"
	dbp_log "Next: make db-patch-gen   (from the same directory)"
}

main "$@"
