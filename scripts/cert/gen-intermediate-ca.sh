#!/usr/bin/env bash
set -e

TYPE=$1

if [[ "$TYPE" != "client-ca" && "$TYPE" != "server-ca" ]]; then
  echo "Invalid type '$TYPE'. Allowed values: client-ca, server-ca." >&2
  exit 1
fi

set -a
CWD="$(dirname "$0")"
PKI_DIR="$CWD/pki/$TYPE"
PROFILE="$CWD/profiles/$TYPE.env.sh"
source $PROFILE
set +a

mkdir -p $PKI_DIR
touch $PKI_DIR/index.txt
echo 1000 > $PKI_DIR/serial.txt

openssl genrsa -out $PKI_DIR/$CN.key 4096

openssl req -new \
  -key $PKI_DIR/$CN.key \
  -out $PKI_DIR/$CN.csr \
  -config $CWD/openssl-intermediate-ca.cnf

openssl ca -batch \
  -config $CWD/openssl-intermediate-ca.cnf \
  -extensions v3_intermediate_ca \
  -days $DAYS \
  -name root_ca \
  -in $PKI_DIR/$CN.csr \
  -out $PKI_DIR/$CN.crt \
  -cert $CWD/pki/root-ca/$CA_CN.crt \
  -keyfile $CWD/pki/root-ca/$CA_CN.key
  