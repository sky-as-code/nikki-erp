#!/usr/bin/env bash
set -e

set -a
PKI_DIR="$CWD/pki/client-cert"
PROFILE="$CWD/profiles/client-cert.env.sh"
source $PROFILE
set +a

mkdir -p $PKI_DIR

openssl genpkey \
  -algorithm RSA \
  -pkeyopt rsa_keygen_bits:2048 \
  -out $PKI_DIR/$CN.key

openssl req -new \
  -key $PKI_DIR/$CN.key \
  -out $PKI_DIR/$CN.csr \
  -config $SDIR/openssl-client-cert.cnf

openssl x509 -req \
  -in $PKI_DIR/$CN.csr \
  -CA $CWD/pki/client-ca/$CA_CN.crt \
  -CAkey $CWD/pki/client-ca/$CA_CN.key \
  -CAcreateserial \
  -out $PKI_DIR/$CN.crt \
  -days $DAYS \
  -extfile $SDIR/openssl-client-cert.cnf \
  -extensions v3_client
