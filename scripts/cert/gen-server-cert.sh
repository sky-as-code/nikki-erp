#!/usr/bin/env bash
set -e

if [[ -n "${APP_ENV:-}" ]]; then
  ENV_SUFFIX="-$APP_ENV"
else
  ENV_SUFFIX=""
fi

set -a
PKI_DIR="$CWD/pki/server-cert"
PROFILE="$CWD/profiles/server-cert$ENV_SUFFIX.env.sh"
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
  -config $SDIR/openssl-server-cert.cnf

openssl x509 -req \
  -in $PKI_DIR/$CN.csr \
  -CA $CWD/pki/server-ca/$CA_CN.crt \
  -CAkey $CWD/pki/server-ca/$CA_CN.key \
  -CAcreateserial \
  -out $PKI_DIR/$CN.crt \
  -days $DAYS \
  -extfile $SDIR/openssl-server-cert.cnf \
  -extensions v3_server

cat $PKI_DIR/$CN.crt $CWD/pki/server-ca/$CA_CN.crt > $PKI_DIR/$CN-chain.crt && \
rm $PKI_DIR/$CN.crt && \
mv $PKI_DIR/$CN-chain.crt $PKI_DIR/$CN.crt
