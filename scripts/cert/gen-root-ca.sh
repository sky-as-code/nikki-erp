#!/usr/bin/env bash
set -e

CWD="$(dirname "$0")"
PKI_DIR="$CWD/pki/root-ca"
PROFILE="$CWD/profiles/root-ca.env.sh"
set -a
source $PROFILE
set +a


mkdir -p $PKI_DIR
touch $PKI_DIR/index.txt
echo 1000 > $PKI_DIR/serial.txt

openssl genrsa -out $PKI_DIR/$CN.key 4096

openssl req -x509 -new -nodes \
  -key $PKI_DIR/$CN.key \
  -days $DAYS \
  -out $PKI_DIR/$CN.crt \
  -config $CWD/openssl-root-ca.cnf \
  -extensions v3_root_ca