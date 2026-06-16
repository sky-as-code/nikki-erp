#!/bin/sh
if [ ! -f /vernemq/etc/vmq.passwd ]; then
  printf "%s\n%s\n" "$VERNE_PASSWORD" "$VERNE_PASSWORD" | vmq-passwd -c /vernemq/etc/vmq.passwd "$VERNE_USER"
  echo "Created Verne user: $VERNE_USER"
fi
exec /usr/sbin/start_vernemq

