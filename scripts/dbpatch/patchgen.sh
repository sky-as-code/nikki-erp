#!/usr/bin/env bash
# Stage 2 of 3: generate the schema patch for human review. Never connects to live: it diffs the
# live_schema.sql snapshot from inspectlive.sh against the tree, both replayed on atlas's dev database.
# Writes patch.sql (nothing that drops) and destructive.review.sql (the drops), and appends to REPORT.md.
# Usage: patchgen.sh <cwd> [run dir]    (run dir defaults to the newest under <cwd>scripts/dbpatch-out)
set -euo pipefail

source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

write_review() {
	local patch="${dbp_run_dir}/patch.sql" destructive="${dbp_run_dir}/destructive.review.sql"
	echo
	echo "## Review"
	echo
	echo "- Generated: $(date '+%Y-%m-%d %H:%M:%S %z')"
	echo "- \`patch.sql\`: $(dbp_count_statements "$patch") statement(s), nothing that drops."
	echo "- \`destructive.review.sql\`: $(dbp_count_statements "$destructive") statement(s), applied only with"
	echo "  \`include_destructive=1\`. An object that exists only in live may hold data: decide per statement."
	echo
	awk -f "${dbp_lib_dir}/review.awk" "$patch" "$destructive"
}

main() {
	[ "$#" -ge 1 ] || dbp_die "usage: $0 <cwd> [run dir]"
	dbp_resolve_tree "$1"
	dbp_resolve_run_dir "${2:-}"
	dbp_require_tools atlas docker sha256sum
	dbp_assert_inspection
	dbp_init_tmp
	local patch="${dbp_run_dir}/patch.sql" destructive="${dbp_run_dir}/destructive.review.sql"
	# A re-run replaces the previous review instead of stacking a second one.
	sed -i '/^## Review$/,$d' "${dbp_run_dir}/REPORT.md"

	dbp_log "Preparing the tree ${dbp_migration_dir}..."
	dbp_prepare_desired
	dbp_log "Diffing live's snapshot against the tree (replays both on a dev database; takes a while)..."
	dbp_atlas_diff patch "$(dbp_file_url "${dbp_run_dir}/live_schema.sql")" "$(dbp_file_url "$dbp_desired_dir")" \
		> "$patch"
	dbp_log "Replaying the patch to find what it leaves out..."
	dbp_residual "$destructive" "${dbp_run_dir}/live_schema.sql" "$patch" \
		|| dbp_die "patch.sql does not replay on live's schema; see the atlas error above."

	if [ ! -s "$patch" ] && [ ! -s "$destructive" ]; then
		printf '\n## Review\n\nLive is up to date with the tree: nothing to patch.\n' >> "${dbp_run_dir}/REPORT.md"
		dbp_log "Live is up to date with the tree: nothing to patch."
		return
	fi
	write_review >> "${dbp_run_dir}/REPORT.md"
	dbp_log "Review these, then run 'make db-patch-apply url=...':"
	dbp_log "  ${patch}"
	dbp_log "  ${destructive}"
	dbp_log "  ${dbp_run_dir}/REPORT.md"
}

main "$@"
