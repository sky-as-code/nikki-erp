# Fallback drift report, for when no commit in history has the atlas.sum live was built from.
# Usage: awk -f drift.awk revisions.tsv <atlas.sum>...
# revisions.tsv: a header, then "version<TAB>hash<TAB>...". Each atlas.sum: the directory hash, then
# "<file> h1:<hash>"; several sums hash the same tree in different line endings, and any of them may match.
# Atlas chains its hashes (each covers every file before it), so only the first divergence is certain:
# every applied file after it differs in hash whether or not its own content changed.

FILENAME == ARGV[1] {
	if (FNR > 1) {
		applied[$1] = $2
		applied_count++
	}
	next
}

FNR == 1 { next }

{
	if (!($1 in hashes)) order[++file_count] = $1
	hashes[$1] = hashes[$1] " " substr($2, 4) " "
}

function add_row(file, status) {
	rows[++row_count] = "| `" file "` | " status " | " \
		((file ~ /_schema\.sql$/) ? "schema: covered by the patch" : "not covered: re-run by hand if needed") " |"
}

function classify(file,    version) {
	version = file
	sub(/_.*/, "", version)
	seen[version] = 1
	if (!(version in applied)) add_row(file, "new")
	else if (diverged) add_row(file, "unknown: after the first divergence")
	else if (index(hashes[file], " " applied[version] " ") == 0) {
		diverged = 1
		add_row(file, "**first divergence**: edited after apply")
	} else unchanged++
}

END {
	for (i = 1; i <= file_count; i++) classify(order[i])
	for (version in applied) {
		if (!(version in seen)) rows[++row_count] = "| version `" version "` | applied, no longer in the tree | - |"
	}
	if (applied_count == 0) {
		print "Live has no `atlas_schema_revisions` rows: it was not built with `atlas migrate apply`, so every file"
		print "counts as new."
	} else {
		print "No commit in history has the atlas.sum live was built from, so per-file drift is known only up to"
		print "the first divergence. " (unchanged + 0) " file(s) before it are unchanged."
	}
	if (row_count == 0) exit
	print ""
	print "| File | Status | Coverage |"
	print "|---|---|---|"
	for (i = 1; i <= row_count; i++) print rows[i]
}
