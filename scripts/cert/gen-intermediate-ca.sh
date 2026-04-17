#!/usr/bin/env bash
set -e

TYPE=$1

if [[ "$TYPE" != "client-ca" && "$TYPE" != "server-ca" ]]; then
  echo "Invalid type '$TYPE'. Allowed values: client-ca, server-ca." >&2
  exit 1
fi

set -a
PKI_DIR="$CWD/pki/$TYPE"
PROFILE="$CWD/profiles/$TYPE.env.sh"
source $PROFILE
set +a

mkdir -p $PKI_DIR

openssl genpkey \
  -algorithm RSA \
  -pkeyopt rsa_keygen_bits:4096 \
  -out $PKI_DIR/$CN.key

openssl req -new \
  -key $PKI_DIR/$CN.key \
  -out $PKI_DIR/$CN.csr \
  -config $SDIR/openssl-intermediate-ca.cnf

openssl x509 -req \
  -in $PKI_DIR/$CN.csr \
  -CA $CWD/pki/root-ca/$CA_CN.crt \
  -CAkey $CWD/pki/root-ca/$CA_CN.key \
  -CAcreateserial \
  -out $PKI_DIR/$CN.crt \
  -days $DAYS \
  -extfile $SDIR/openssl-intermediate-ca.cnf \
  -extensions v3_intermediate_ca