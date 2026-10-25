#!/usr/bin/env bash
# Applies every seed in this folder, in name order, as ONE transaction: any failure leaves nothing behind.
# Several seeds are plain INSERTs, so a second run fails and rolls back; rebuild with `make db-infra-apply`.
set -euo pipefail
cd "$(dirname "$0")"
export LC_ALL=C
shopt -s nullglob

args=()
for file in *.sql; do
	args+=(-f "$file")
done
if [ ${#args[@]} -eq 0 ]; then
	echo "No seed files in $(pwd)"
	exit 0
fi

PGPASSWORD="$POSTGRES_PASSWORD" psql -U "$POSTGRES_USER" -d nikki_erp \
	-v ON_ERROR_STOP=1 --single-transaction "${args[@]}"
