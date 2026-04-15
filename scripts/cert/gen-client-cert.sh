#!/usr/bin/env bash
set -e

set -a
CWD="$(dirname "$0")"
PKI_DIR="$CWD/pki/client-cert"
PROFILE="$CWD/profiles/client-cert.env.sh"
source $PROFILE
set +a

mkdir -p $PKI_DIR

openssl genrsa -out $PKI_DIR/$CN.key 2048

openssl req -new \
  -key $PKI_DIR/$CN.key \
  -out $PKI_DIR/$CN.csr \
  -config $CWD/openssl-client-cert.cnf

openssl ca -batch \
  -config $CWD/openssl-client-cert.cnf \
  -extensions v3_client \
  -days $DAYS \
  -name client_ca \
  -in $PKI_DIR/$CN.csr \
  -out $PKI_DIR/$CN.crt \
  -cert $CWD/pki/client-ca/$CA_CN.crt \
  -keyfile $CWD/pki/client-ca/$CA_CN.key
