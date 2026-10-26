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
# Must match CORE.S3_STORAGE.BUCKET in config/config.default.yaml.
MINIO_BUCKET="nikki-local"

echo "Waiting for MinIO to be ready..."
for i in $(seq 1 30); do
	if docker exec nikki_minio mc alias set "$MINIO_ALIAS" http://localhost:9000 "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD" >/dev/null 2>&1; then
		break
	fi
	sleep 1
done

echo "Creating MinIO bucket ${MINIO_BUCKET}..."
docker exec nikki_minio mc mb --ignore-existing "${MINIO_ALIAS}/${MINIO_BUCKET}"

# IT Storage tags an upload itstorage-state=pending until a record references it, and again once a
# record drops it; this rule expires such objects after 24 hours (1 day is the rule's minimum unit).
# Must match constants.ObjectTagState. A deployment on another bucket needs the same rule.
MINIO_PENDING_TAG="itstorage-state=pending"
if docker exec nikki_minio mc ilm rule export "${MINIO_ALIAS}/${MINIO_BUCKET}" 2>/dev/null | grep -q '"itstorage-state"'; then
	echo "MinIO lifecycle rule for ${MINIO_PENDING_TAG} already exists"
else
	echo "Adding MinIO lifecycle rule: expire ${MINIO_PENDING_TAG} objects after 1 day..."
	docker exec nikki_minio mc ilm rule add --expire-days 1 --tags "$MINIO_PENDING_TAG" "${MINIO_ALIAS}/${MINIO_BUCKET}"
fi

echo "Creating MinIO access token..."
CREDENTIALS_JSON=$(docker exec nikki_minio mc admin accesskey create "$MINIO_ALIAS" --json)

ACCESS_KEY=$(echo "$CREDENTIALS_JSON" | grep -o '"accessKey":"[^"]*"' | cut -d'"' -f4)
SECRET_KEY=$(echo "$CREDENTIALS_JSON" | grep -o '"secretKey":"[^"]*"' | cut -d'"' -f4)

if [ -z "$ACCESS_KEY" ] || [ -z "$SECRET_KEY" ]; then
	echo "Error: failed to parse MinIO access token from mc output" >&2
	exit 1
fi

env_file="${config_dir}/.env.local"
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
