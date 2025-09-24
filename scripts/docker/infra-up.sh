#!/bin/bash

# Exit immediately if any command fails
set -e

# Check if CWD parameter is provided
if [ -z "$1" ]; then
    echo "Error: CWD parameter is required. Please provide the current working directory path."
    echo "Usage: $0 <cwd_path>"
    exit 1
fi

# Set CWD environment variable from first parameter
export CWD="$1"

echo "CWD: $CWD"

# Function to handle errors
handle_error() {
    echo "Error: An error occurred during script execution. Exiting without starting infrastructure services."
    exit 1
}

# Set trap to catch errors
trap handle_error ERR

# Check if the current machine is part of a Docker Swarm, if not, initialize it
# if ! docker info --format '{{.Swarm.LocalNodeState}}' 2>/dev/null | grep -q 'active'; then
if [ "$(docker info --format '{{.Swarm.LocalNodeState}}' 2>/dev/null)" = "inactive" ]; then
    echo "Docker Swarm not initialized. Initializing Docker Swarm..."
    docker swarm init --advertise-addr 127.0.0.1
else
    echo "Docker Swarm already initialized."
fi

# Function to create secret if it doesn't exist
create_secret_if_not_exists() {
    local secret_name="$1"
    local secret_value="$2"
    
    if docker secret ls --format "{{.Name}}" | grep -q "^${secret_name}$"; then
        echo "Secret '${secret_name}' already exists, skipping creation."
    else
        echo "Creating secret '${secret_name}'..."
        echo "${secret_value}" | docker secret create "${secret_name}" -
    fi
}

# Create Docker secrets for all usernames and passwords
echo "Creating Docker secrets..."

# PostgreSQL secrets
create_secret_if_not_exists "postgres_username" "nikki_admin"
create_secret_if_not_exists "postgres_password" "nikki_password"

# PgAdmin secrets
create_secret_if_not_exists "pgadmin_password" "admin"

# KeyDB secrets
create_secret_if_not_exists "keydb_password" "nikki_password"

# RabbitMQ secrets
create_secret_if_not_exists "rabbitmq_username" "nikki_admin"
create_secret_if_not_exists "rabbitmq_password" "nikki_password"

echo "Docker secrets check completed!"

# Start infrastructure services
echo "Starting infrastructure services..."
# docker compose -f "$(dirname "$0")/docker-compose.local.yml" up -d
docker stack deploy -c "$(dirname "$0")/docker-compose.local.yml" --detach=false nikki_infra
