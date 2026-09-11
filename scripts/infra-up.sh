#!/usr/bin/env bash
set -euo pipefail

if [ -z "${1:-}" ] || [ -z "${2:-}" ]; then
	echo "Error: compose file and config directory parameters are required. Usage: $0 <compose_file> <config_dir>" >&2
	exit 1
fi

compose_file="$1"
config_dir="$2"

docker compose -f "$compose_file" up -d

MINIO_ROOT_USER="nikki_admin"
MINIO_ROOT_PASSWORD="nikki_password"
MINIO_ALIAS="local"

echo "Waiting for MinIO to be ready..."
for i in $(seq 1 30); do
	if docker exec nikki_minio mc alias set "$MINIO_ALIAS" http://localhost:9000 "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD" >/dev/null 2>&1; then
		break
	fi
	sleep 1
done

echo "Creating MinIO access token..."
CREDENTIALS_JSON=$(docker exec nikki_minio mc admin accesskey create "$MINIO_ALIAS" --json)

ACCESS_KEY=$(echo "$CREDENTIALS_JSON" | grep -o '"accessKey":"[^"]*"' | cut -d'"' -f4)
SECRET_KEY=$(echo "$CREDENTIALS_JSON" | grep -o '"secretKey":"[^"]*"' | cut -d'"' -f4)

if [ -z "$ACCESS_KEY" ] || [ -z "$SECRET_KEY" ]; then
	echo "Error: failed to parse MinIO access token from mc output" >&2
	exit 1
fi

env_file="${config_dir}/local.env"
sample_file="${env_file}.sample"

if [ ! -f "$env_file" ]; then
	echo "Creating ${env_file} from ${sample_file}..."
	cp "$sample_file" "$env_file"
fi

if grep -q '^CORE_S3_STORAGE_ACCESS_TOKEN=' "$env_file"; then
	sed -i "s|^CORE_S3_STORAGE_ACCESS_TOKEN=.*|CORE_S3_STORAGE_ACCESS_TOKEN=${ACCESS_KEY}|" "$env_file"
else
	printf '\nCORE_S3_STORAGE_ACCESS_TOKEN=%s\n' "$ACCESS_KEY" >> "$env_file"
fi

if grep -q '^CORE_S3_STORAGE_SECRET_KEY=' "$env_file"; then
	sed -i "s|^CORE_S3_STORAGE_SECRET_KEY=.*|CORE_S3_STORAGE_SECRET_KEY=${SECRET_KEY}|" "$env_file"
else
	printf 'CORE_S3_STORAGE_SECRET_KEY=%s\n' "$SECRET_KEY" >> "$env_file"
fi

echo "Wrote S3/MinIO access token to ${env_file}"
