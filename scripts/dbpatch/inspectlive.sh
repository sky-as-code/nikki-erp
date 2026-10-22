#!/usr/bin/env bash
# Stage 1 of 3: snapshot the public schema of a running database, read-only (catalog queries only, no rows,
# no locks). Writes live_schema.sql, revisions.tsv, inspect.env and REPORT.md to <cwd>scripts/dbpatch-out/<ts>/.
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

write_revisions() {
	local schema="$1"
	printf 'version\thash\tapplied\ttotal\texecuted_at\n'
	[ -n "$schema" ] || return 0
	dbp_psql "$dbp_url" -At -F $'\t' -c "
		SELECT version, hash, applied, total, executed_at FROM \"${schema}\".atlas_schema_revisions
		WHERE version ~ '^[0-9]' ORDER BY version"
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

dbp_git() {
	git -C "$dbp_migration_dir" "$@"
}

# Usage: tree_has_hash <commit> <hash>. The committed atlas.sum is often stale, because the tree is re-hashed
# only right before an apply, so the commit's files are hashed again too, in both line endings.
tree_has_hash() {
	local dir="${dbp_tmp}/history"
	dbp_git show "${1}:./atlas.sum" 2>/dev/null | grep -qF "h1:${2}" && return 0
	rm -rf "$dir"
	mkdir -p "${dir}/src"
	dbp_git archive "$1" . | tar -x -C "${dir}/src"
	hash_tree_variants "${dir}/src" "$dir"
	cat "${dir}/lf/atlas.sum" "${dir}/crlf/atlas.sum" | grep -qF "h1:${2}"
}

# Atlas chains its hashes: a file's hash covers every file before it. The hash of the last applied revision
# therefore names the exact tree live was built from, and the commit holding it gives a per-file diff.
find_source_commit() {
	local hash commit
	hash="$(tail -n +2 "$1" | tail -n 1 | cut -f2)"
	[ -n "$hash" ] || return 0
	while read -r commit; do
		if tree_has_hash "$commit" "$hash"; then
			echo "$commit"
			return 0
		fi
	done < <(dbp_git log --format=%H -n 200 -- . 2>/dev/null)
}

coverage_of() {
	case "$1" in
		*_schema.sql) echo "schema: covered by the patch" ;;
		*) echo "not covered: re-run by hand if needed" ;;
	esac
}

# Compares with the working tree, uncommitted and untracked files included.
write_commit_drift() {
	local commit="$1" status file changes
	dbp_git log -1 --date=short --format='Live was built from the tree at commit `%h` (%ad, "%s").' "$commit"
	echo
	changes="$(
		dbp_git diff --no-renames --name-status "$commit" -- '*.sql'
		dbp_git ls-files --others --exclude-standard -- '*.sql' | sed 's/^/A\t/'
	)"
	if [ -z "$changes" ]; then
		echo "No migration file changed since."
		return
	fi
	echo "| File | Status | Coverage |"
	echo "|---|---|---|"
	echo "$changes" | sort -k2 | while IFS=$'\t' read -r status file; do
		case "$status" in
			M) status="edited after apply" ;;
			A) status="new" ;;
			D) status="removed from the tree" ;;
		esac
		echo "| \`${file##*/}\` | ${status} | $(coverage_of "$file") |"
	done
}

# Usage: hash_tree_variants <source dir> <out dir>. Live recorded the hashes of the files as they were on the
# machine that applied them, and line endings differ between checkouts (core.autocrlf), so the tree is hashed
# in both forms, into <out dir>/lf and <out dir>/crlf.
hash_tree_variants() {
	local variant dir file
	for variant in lf crlf; do
		dir="${2}/${variant}"
		mkdir -p "$dir"
		for file in "$1"/*.sql; do
			# Only real line breaks: a last line without one stays as it is, as in a checkout.
			if [ "$variant" = lf ]; then perl -pe 's/\r?\n/\n/' "$file"; else perl -pe 's/\r?\n/\r\n/' "$file"; fi \
				> "${dir}/${file##*/}"
		done
		dbp_hash_dir "$dir"
	done
}

write_drift() {
	local revisions="$1" commit
	commit="$(find_source_commit "$revisions")"
	if [ -n "$commit" ]; then
		write_commit_drift "$commit"
		return
	fi
	hash_tree_variants "$dbp_migration_dir" "$dbp_tmp"
	awk -f "${dbp_lib_dir}/drift.awk" "$revisions" "${dbp_tmp}/lf/atlas.sum" "${dbp_tmp}/crlf/atlas.sum"
}

write_report() {
	local run_dir="$1" env_file="$1/inspect.env"
	echo "# DB patch report"
	echo
	echo "- Live: \`${dbp_url_label}\`, PostgreSQL $(dbp_env_get "$env_file" server_version)"
	echo "- Tree: \`${dbp_migration_dir}\`"
	echo "- Inspected: $(dbp_env_get "$env_file" inspected_at)"
	echo
	echo "## Drift"
	echo
	write_drift "${run_dir}/revisions.tsv"
}

main() {
	[ "$#" -ge 1 ] || dbp_die "usage: $0 <cwd> [url]"
	dbp_resolve_tree "$1"
	dbp_resolve_url "${2:-}"
	dbp_require_tools atlas docker sha256sum git perl
	dbp_init_tmp
	local run_dir schema
	run_dir="${dbp_out_root}/$(date +%Y%m%d-%H%M%S)"
	mkdir -p "$run_dir"

	dbp_log "Inspecting the public schema of ${dbp_url_label} (read-only)..."
	dbp_atlas_inspect "$dbp_url" > "${run_dir}/live_schema.sql"
	[ -s "${run_dir}/live_schema.sql" ] || dbp_die "atlas returned an empty schema for ${dbp_url_label}."
	schema="$(find_revisions_schema)"
	write_revisions "$schema" > "${run_dir}/revisions.tsv"
	write_inspect_env "$run_dir" "$schema" > "${run_dir}/inspect.env"
	write_report "$run_dir" > "${run_dir}/REPORT.md"

	dbp_log "Inspection written to ${run_dir}"
	dbp_log "Next: make db-patch-gen   (from the same directory)"
}

main "$@"
