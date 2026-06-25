#!/usr/bin/env bash
set -e

PKI_DIR="$CWD/pki/root-ca"
PROFILE="$CWD/profiles/root-ca.env.sh"
set -a
source $PROFILE
set +a

mkdir -p $PKI_DIR

openssl genpkey \
  -algorithm RSA \
  -pkeyopt rsa_keygen_bits:4096 \
  -out $PKI_DIR/$CN.key

openssl req -x509 -new -nodes \
  -key $PKI_DIR/$CN.key \
  -days $DAYS \
  -out $PKI_DIR/$CN.crt \
  -config $SDIR/openssl-root-ca.cnf \
  -extensions v3_root_ca