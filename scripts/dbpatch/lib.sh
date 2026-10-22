#!/usr/bin/env bash
# Shared by inspectlive.sh, patchgen.sh and patchapply.sh. Source it, do not run it.

dbp_lib_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
dbp_pg_image="postgres:17"
# A pattern is "<schema>.<object>" even with --schema set. Functions, triggers and views need no pattern:
# the community build of atlas skips them on its own, and this tool leaves them out of scope anyway.
dbp_excludes=(--exclude '*.atlas_schema_revisions')
dbp_inspect_files=(live_schema.sql revisions.tsv inspect.env)

dbp_die() {
	echo "Error: $*" >&2
	exit 1
}

dbp_log() {
	echo "==> $*" >&2
}

dbp_require_tools() {
	local tool
	for tool in "$@"; do
		command -v "$tool" >/dev/null 2>&1 || dbp_die "'${tool}' is not on PATH."
	done
}

# Atlas is a native Windows binary under Git Bash: it wants "F:/..." in a file:// URL, not "/f/...".
dbp_file_url() {
	if command -v cygpath >/dev/null 2>&1; then
		echo "file://$(cygpath -m "$1")"
	else
		echo "file://$1"
	fi
}

dbp_sha() {
	sha256sum "$1" | cut -d' ' -f1
}

dbp_env_get() {
	grep -m1 "^$2=" "$1" | cut -d= -f2- || true
}

# Counts statements, not lines: atlas spreads a CREATE TABLE over several lines.
dbp_count_statements() {
	if [ -s "$1" ]; then grep -cE ';[[:space:]]*$' "$1" || true; else echo 0; fi
}

dbp_init_tmp() {
	dbp_tmp="$(mktemp -d)"
	trap 'rm -rf "$dbp_tmp"' EXIT
}

# <cwd> is the directory `make` ran from; it picks the tree: <cwd>scripts/migrations.
dbp_resolve_tree() {
	local cwd="$1"
	[ -d "${cwd}scripts/migrations" ] || dbp_die "no migrations directory at '${cwd}scripts/migrations'."
	dbp_migration_dir="$(cd "${cwd}scripts/migrations" && pwd)"
	dbp_out_root="$(cd "${cwd}scripts" && pwd)/dbpatch-out"
	dbp_atlas_hcl="$(dbp_file_url "${dbp_lib_dir}/atlas.hcl")"
}

# The URL holds credentials: it comes from the argument or DBPATCH_URL and is never echoed or written to disk.
dbp_resolve_url() {
	dbp_url="${1:-${DBPATCH_URL:-}}"
	[ -n "$dbp_url" ] || dbp_die "the live database URL is required: url=... or DBPATCH_URL."
	local pattern='^postgres(ql)?://([^@/]*@)?([^/:?]+)(:([0-9]+))?/([^?]+)'
	[[ "$dbp_url" =~ $pattern ]] || dbp_die "url must look like postgres://user:pass@host[:port]/dbname[?params]."
	dbp_url_host="${BASH_REMATCH[3]}"
	dbp_url_port="${BASH_REMATCH[5]:-5432}"
	dbp_url_db="${BASH_REMATCH[6]}"
	dbp_url_label="${dbp_url_host}:${dbp_url_port}/${dbp_url_db}"
}

# dir=<path> when given, otherwise the newest run under dbpatch-out.
dbp_resolve_run_dir() {
	local dir="${1:-}"
	if [ -z "$dir" ]; then
		dir="$(find "$dbp_out_root" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | sort | tail -n 1)"
		[ -n "$dir" ] || dbp_die "no inspection under '${dbp_out_root}'; run 'make db-inspect-live' first."
	fi
	[ -d "$dir" ] || dbp_die "'${dir}' is not a directory."
	dbp_run_dir="$(cd "$dir" && pwd)"
}

dbp_assert_inspection() {
	local name env_file="${dbp_run_dir}/inspect.env"
	for name in "${dbp_inspect_files[@]}"; do
		[ -s "${dbp_run_dir}/${name}" ] \
			|| dbp_die "'${dbp_run_dir}/${name}' is missing or empty; run 'make db-inspect-live' again."
	done
	[ "$(dbp_env_get "$env_file" tree)" = "$dbp_migration_dir" ] \
		|| dbp_die "the inspection was made for tree '$(dbp_env_get "$env_file" tree)', not '${dbp_migration_dir}'."
	[ "$(dbp_sha "${dbp_run_dir}/live_schema.sql")" = "$(dbp_env_get "$env_file" live_schema_sha256)" ] \
		|| dbp_die "live_schema.sql changed after the inspection; run 'make db-inspect-live' again."
}

# libpq rejects atlas-only URL parameters such as search_path. A psql container reaches a database on
# this machine through host.docker.internal.
dbp_psql() {
	local url="$1"
	shift
	url="$(echo "$url" | sed -E 's/([?&])search_path=[^&]*&?/\1/; s/[?&]$//')"
	if command -v psql >/dev/null 2>&1; then
		psql "$url" -X -q "$@"
		return
	fi
	url="$(echo "$url" | sed -E 's#^([a-z]+://([^@/]*@)?)(localhost|127\.0\.0\.1)([:/])#\1host.docker.internal\4#')"
	docker run --rm -i --add-host=host.docker.internal:host-gateway "$dbp_pg_image" psql "$url" -X -q "$@"
}

# Atlas prints a Pro upsell on stderr for every schema command; stderr is shown only when it fails. A failed
# replay quotes the whole statement, megabytes for a seed file: only its start and the verdict at the end.
dbp_atlas() {
	local log
	log="$(mktemp)"
	if ! atlas "$@" 2>"$log"; then
		if [ "$(wc -c < "$log")" -gt 4000 ]; then
			{ head -c 1000 "$log"; printf '\n[...]\n'; tail -c 2000 "$log"; } >&2
		else
			cat "$log" >&2
		fi
		rm -f "$log"
		return 1
	fi
	rm -f "$log"
}

dbp_atlas_inspect() {
	dbp_atlas schema inspect --url "$1" --schema public "${dbp_excludes[@]}" --format '{{ sql . "  " }}'
}

# Usage: dbp_atlas_diff <env> <from url> <to url>. Both sides are replayed on atlas's dev database.
dbp_atlas_diff() {
	dbp_atlas schema diff --config "$dbp_atlas_hcl" --env "$1" --from "$2" --to "$3" --schema public \
		"${dbp_excludes[@]}" --format '{{ sql . "  " }}'
}

dbp_hash_dir() {
	dbp_atlas migrate hash --dir "$(dbp_file_url "$1")"
}

# A hashed copy of the tree: the repo's atlas.sum is never rewritten, and a stale one cannot fail the run.
dbp_prepare_desired() {
	dbp_desired_dir="${dbp_tmp}/desired"
	mkdir -p "$dbp_desired_dir"
	cp "$dbp_migration_dir"/*.sql "$dbp_desired_dir/"
	dbp_hash_dir "$dbp_desired_dir"
}

# Usage: dbp_residual <out file> <sql file>... Replays the files in order on the dev database and writes
# what the tree still wants on top of them. Replaying is also the rehearsal: a bad statement fails here.
dbp_residual() {
	local out="$1" dir="${dbp_tmp}/rehearse" seq=1 file
	shift
	rm -rf "$dir"
	mkdir -p "$dir"
	for file in "$@"; do
		[ -s "$file" ] || continue
		cp "$file" "${dir}/${seq}_step.sql"
		seq=$((seq + 1))
	done
	dbp_hash_dir "$dir"
	dbp_atlas_diff full "$(dbp_file_url "$dir")" "$(dbp_file_url "$dbp_desired_dir")" > "$out"
}
