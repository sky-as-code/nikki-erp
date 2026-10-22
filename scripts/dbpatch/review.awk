# Flags statements that can fail on a table that already has rows, take heavy locks, or lose data.
# Usage: awk -f review.awk patch.sql destructive.review.sql
# Static only: nothing here reads live data, so every flag is a "check before applying", not a verdict.

# Splits an ALTER TABLE body into its clauses on top-level commas; numeric(10,2) and quoted names stay whole.
function split_clauses(body, out,    i, ch, depth, quote, start, count) {
	depth = 0; quote = ""; start = 1; count = 0
	for (i = 1; i <= length(body); i++) {
		ch = substr(body, i, 1)
		if (quote != "") { if (ch == quote) quote = ""; continue }
		if (ch == "'" || ch == "\"") quote = ch
		else if (ch == "(") depth++
		else if (ch == ")") depth--
		else if (ch == "," && depth == 0) { out[++count] = substr(body, start, i - start); start = i + 1 }
	}
	out[++count] = substr(body, start)
	return count
}

function table_of(stmt,    name) {
	if (match(stmt, /(ALTER TABLE|ON) "[^"]+"/)) {
		name = substr(stmt, RSTART, RLENGTH)
		sub(/^(ALTER TABLE|ON) /, "", name)
		return name
	}
	return "?"
}

function flag(table, clause, risk) {
	gsub(/^ +|;? *$/, "", clause)
	if (length(clause) > 110) clause = substr(clause, 1, 107) "..."
	warnings[++warning_count] = "| " table " | `" clause "` | " risk " |"
}

function review_clause(table, clause) {
	if (clause ~ /ADD COLUMN/ && clause ~ /NOT NULL/ && clause !~ /DEFAULT/)
		flag(table, clause, "NOT NULL without a default: fails if the table has rows")
	else if (clause ~ /SET NOT NULL/)
		flag(table, clause, "fails if the column holds NULLs")
	else if (clause ~ /ALTER COLUMN .* TYPE /)
		flag(table, clause, "type change: rewrites the table under an exclusive lock; fails on values that do not cast")
	else if (clause ~ /ADD CONSTRAINT .* UNIQUE/)
		flag(table, clause, "fails on duplicate values")
	else if (clause ~ /ADD CONSTRAINT .* FOREIGN KEY/)
		flag(table, clause, "fails on orphan rows; scans the table")
	else if (clause ~ /ADD CONSTRAINT .* CHECK/)
		flag(table, clause, "fails on rows that violate the check")
	if (clause ~ /ADD COLUMN/) added[table] = 1
}

function review_patch(stmt,    body, clauses, count, i, table) {
	table = table_of(stmt)
	if (stmt ~ /^CREATE UNIQUE INDEX/) {
		flag(table, stmt, "fails on duplicate values; blocks writes while it builds")
	} else if (stmt ~ /^CREATE INDEX/) {
		flag(table, stmt, "blocks writes while it builds: on a big table, build it CONCURRENTLY by hand instead")
	} else if (stmt ~ /^ALTER TABLE/) {
		body = stmt
		sub(/^ALTER TABLE "[^"]+" /, "", body)
		count = split_clauses(body, clauses)
		for (i = 1; i <= count; i++) review_clause(table, clauses[i])
	}
}

function review_destructive(stmt,    table) {
	table = table_of(stmt)
	if (stmt ~ /DROP COLUMN/) dropped[table] = 1
}

# Joins atlas's multi-line statements; comment lines are skipped.
/^--/ || /^[[:space:]]*$/ { next }

{
	line = $0
	sub(/^[[:space:]]+/, "", line)
	stmt = (stmt == "") ? line : stmt " " line
	if (line !~ /;[[:space:]]*$/) next
	if (FILENAME == ARGV[1]) review_patch(stmt)
	else review_destructive(stmt)
	stmt = ""
}

END {
	print "### Warnings"
	print ""
	if (warning_count == 0) {
		print "None: no statement in patch.sql depends on the rows already in live."
	} else {
		print "Check these against live's data before applying. `psql -1` rolls the whole patch back if one fails."
		print ""
		print "| Table | Statement | Risk |"
		print "|---|---|---|"
		for (i = 1; i <= warning_count; i++) print warnings[i]
	}
	for (table in dropped) if (table in added) renames[++rename_count] = table
	if (rename_count == 0) exit
	print ""
	print "### Possible renames"
	print ""
	print "These tables gain columns in patch.sql and lose columns in destructive.review.sql. If a column was"
	print "renamed, copy its data across by hand after patch.sql and before the destructive file, or it is lost."
	print ""
	for (i = 1; i <= rename_count; i++) print "- " renames[i]
}
