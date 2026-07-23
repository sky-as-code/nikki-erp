#!/bin/bash

set -e

VERNE_PASSWD_CONFIG_FILE="/vernemq/etc/vmq.passwd"

VERNE_PASSWORD=$(cat "$VERNE_PASSWORD_FILE")

if [ ! -f $VERNE_PASSWD_CONFIG_FILE ]; then
	printf "%s\n%s\n" "$VERNE_PASSWORD" "$VERNE_PASSWORD" | vmq-passwd -c $VERNE_PASSWD_CONFIG_FILE "$VERNE_USER"
	echo "Created user: $VERNE_USER"
fi

exec /usr/sbin/start_vernemq
