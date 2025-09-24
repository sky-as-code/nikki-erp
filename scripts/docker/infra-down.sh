#!/bin/bash

# Exit immediately if any command fails
set -e

# Function to handle errors
handle_error() {
    echo "Error: An error occurred during cleanup. Some resources may not have been properly removed."
    echo "You may need to manually clean up remaining resources."
    exit 1
}

# Set trap to catch errors
trap handle_error ERR

# Function to check if a secret exists
secret_exists() {
    local secret_name="$1"
    docker secret ls --format "{{.Name}}" | grep -q "^${secret_name}$"
}

# Function to remove secret if it exists
remove_secret_if_exists() {
    local secret_name="$1"
    
    if secret_exists "${secret_name}"; then
        echo "Removing secret '${secret_name}'..."
        docker secret rm "${secret_name}"
    else
        echo "Secret '${secret_name}' does not exist, skipping removal."
    fi
}

# Function to check if stack exists
stack_exists() {
    docker stack ls --format "{{.Name}}" | grep -q "^nikki_infra$"
}

echo "Starting infrastructure cleanup..."

# Remove Docker stack if it exists
if stack_exists; then
    echo "Removing Docker stack 'nikki_infra'..."
    docker stack rm nikki_infra
    
    # Wait for stack removal to complete
    echo "Waiting for stack removal to complete..."
    while stack_exists; do
        echo "Stack still exists, waiting..."
        sleep 2
    done
    echo "Docker stack 'nikki_infra' removed successfully."
else
    echo "Docker stack 'nikki_infra' does not exist, skipping removal."
fi

# Remove Docker secrets
echo "Removing Docker secrets..."

# PostgreSQL secrets
remove_secret_if_exists "postgres_username"
remove_secret_if_exists "postgres_password"

# PgAdmin secrets
remove_secret_if_exists "pgadmin_password"

# KeyDB secrets
remove_secret_if_exists "keydb_password"

# RabbitMQ secrets
remove_secret_if_exists "rabbitmq_username"
remove_secret_if_exists "rabbitmq_password"

echo "Docker secrets cleanup completed!"

echo "Infrastructure cleanup completed successfully!"
