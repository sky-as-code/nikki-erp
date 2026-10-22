#!/usr/bin/env bash
# Stage 3 of 3: apply the reviewed patch to the live database. The only stage that writes to live.
# Order: drift guard -> rehearsal on live's schema (no data) -> typed confirmation -> one transaction ->
# verify -> `atlas migrate set`, so later deploys of this database take the normal `migrate apply` path.
# Usage: patchapply.sh <cwd> [url] [run dir]    (url defaults to $DBPATCH_URL)
# Flags (environment): DBPATCH_DRYRUN=1 stops after the rehearsal; DBPATCH_INCLUDE_DESTRUCTIVE=1 also applies
# destructive.review.sql; DBPATCH_CONFIRM=yes skips the prompt; DBPATCH_ALLOW_RESIDUAL=1 accepts a patch
# that was trimmed during review and so no longer brings live fully up to the tree.
set -euo pipefail

source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

select_patch_files() {
	local patch="${dbp_run_dir}/patch.sql" destructive="${dbp_run_dir}/destructive.review.sql"
	[ -f "$patch" ] || dbp_die "no patch.sql in ${dbp_run_dir}; run 'make db-patch-gen' first."
	dbp_patch_files=("$patch")
	dbp_expected_residual="$destructive"
	if [ "${DBPATCH_INCLUDE_DESTRUCTIVE:-}" = 1 ]; then
		[ -f "$destructive" ] || dbp_die "include_destructive=1, but there is no destructive.review.sql."
		dbp_patch_files+=("$destructive")
		dbp_expected_residual="${dbp_tmp}/empty.sql"
		: > "$dbp_expected_residual"
	fi
	[ -n "$(cat "${dbp_patch_files[@]}")" ] || dbp_die "the patch is empty: live is already up to date."
}

guard_drift() {
	dbp_log "Checking that live has not changed since the inspection..."
	dbp_atlas_inspect "$dbp_url" > "${dbp_tmp}/live_now.sql"
	[ "$(dbp_sha "${dbp_tmp}/live_now.sql")" = "$(dbp_env_get "${dbp_run_dir}/inspect.env" live_schema_sha256)" ] \
		|| dbp_die "live's schema changed after the inspection; start over with 'make db-inspect-live'."
}

# Usage: assert_residual <actual residual file> <what it is checked after>
assert_residual() {
	cmp -s "$1" "$dbp_expected_residual" && return 0
	echo "--- still missing after ${2} (expected: $(basename "$dbp_expected_residual")):" >&2
	cat "$1" >&2
	[ "${DBPATCH_ALLOW_RESIDUAL:-}" = 1 ] && return 0
	return 1
}

rehearse() {
	dbp_log "Rehearsing the patch on live's schema (no data) in a dev database..."
	dbp_residual "${dbp_tmp}/residual.sql" "${dbp_run_dir}/live_schema.sql" "${dbp_patch_files[@]}" \
		|| dbp_die "the patch does not replay on live's schema; see the atlas error above."
	assert_residual "${dbp_tmp}/residual.sql" "the rehearsal" \
		|| dbp_die "the patch no longer matches the tree; regenerate it or pass allow_residual=1."
}

confirm_apply() {
	local answer="" count
	count="$(cat "${dbp_patch_files[@]}" | grep -cE ';[[:space:]]*$' || true)"
	echo "About to apply ${count} statement(s) from: ${dbp_patch_files[*]##*/}" >&2
	echo "to ${dbp_url_label}, in one transaction." >&2
	[ "${DBPATCH_CONFIRM:-}" = yes ] && return 0
	read -r -p "Type 'yes' to continue: " answer || true
	[ "$answer" = yes ] || dbp_die "aborted; nothing was applied."
}

apply_patch() {
	dbp_log "Applying to ${dbp_url_label}..."
	cat "${dbp_patch_files[@]}" | dbp_psql "$dbp_url" -1 -v ON_ERROR_STOP=1 -f - \
		|| dbp_die "the patch failed and was rolled back; live is unchanged. Check REPORT.md §Warnings."
}

verify_live() {
	dbp_log "Verifying live against the tree..."
	dbp_atlas_inspect "$dbp_url" > "${dbp_tmp}/live_after.sql"
	dbp_atlas_diff full "$(dbp_file_url "${dbp_tmp}/live_after.sql")" "$(dbp_file_url "$dbp_desired_dir")" \
		> "${dbp_tmp}/verify.sql"
	assert_residual "${dbp_tmp}/verify.sql" "the apply" \
		|| dbp_die "the patch is COMMITTED, but live still differs from the tree (above); revisions were not reset."
}

# `migrate set` adds and removes rows but keeps the hash of a row that exists, which would keep naming the
# old tree; the next inspection would then report drift that is already patched.
refresh_revision_hashes() {
	local schema="$1" values
	local to_values='NR > 1 { v = $1; sub(/_.*/, "", v); q = sprintf("%c", 39)
		printf "%s(%s%s%s, %s%s%s)", (n++ ? ", " : ""), q, v, q, q, substr($2, 4), q }'
	values="$(awk "$to_values" "${dbp_desired_dir}/atlas.sum")"
	dbp_psql "$dbp_url" -v ON_ERROR_STOP=1 -c "UPDATE \"${schema}\".atlas_schema_revisions AS r SET hash = v.hash
		FROM (VALUES ${values}) AS v(version, hash) WHERE r.version = v.version"
}

# Marks every file of the tree as applied, with its current hash, so `atlas migrate apply` skips them.
rebaseline() {
	local version schema
	version="$(find "$dbp_desired_dir" -maxdepth 1 -name '*.sql' -printf '%f\n' | sort | tail -n 1)"
	version="${version%%_*}"
	schema="$(dbp_env_get "${dbp_run_dir}/inspect.env" revisions_schema)"
	dbp_log "Setting atlas revisions of ${dbp_url_label} to version ${version}..."
	dbp_atlas migrate set "$version" --url "$dbp_url" --dir "$(dbp_file_url "$dbp_desired_dir")" \
		--revisions-schema "${schema:-public}"
	refresh_revision_hashes "${schema:-public}"
	{
		echo
		echo "## Applied"
		echo
		echo "- $(date '+%Y-%m-%d %H:%M:%S %z'): ${dbp_patch_files[*]##*/} applied to \`${dbp_url_label}\`."
		echo "- Atlas revisions set to \`${version}\` in schema \`${schema:-public}\`."
	} >> "${dbp_run_dir}/REPORT.md"
}

main() {
	[ "$#" -ge 1 ] || dbp_die "usage: $0 <cwd> [url] [run dir]"
	dbp_resolve_tree "$1"
	dbp_resolve_url "${2:-}"
	dbp_resolve_run_dir "${3:-}"
	dbp_require_tools atlas docker sha256sum cmp
	dbp_assert_inspection
	dbp_init_tmp
	select_patch_files
	dbp_prepare_desired

	guard_drift
	rehearse
	if [ "${DBPATCH_DRYRUN:-}" = 1 ]; then
		dbp_log "Dry run: the rehearsal passed; live was not touched."
		return
	fi
	confirm_apply
	apply_patch
	verify_live
	rebaseline
	dbp_log "Done. ${dbp_url_label} matches the tree; report: ${dbp_run_dir}/REPORT.md"
}

main "$@"
