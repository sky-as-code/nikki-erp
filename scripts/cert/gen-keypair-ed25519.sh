#!/usr/bin/env bash
set -e

PKI_DIR="$CWD/pki/jwt-keypair"

mkdir -p $PKI_DIR

# Private key (PKCS#8)
openssl genpkey -algorithm Ed25519 -out $PKI_DIR/ed25519.key

# Public key
openssl pkey -in $PKI_DIR/ed25519.key -pubout -out $PKI_DIR/ed25519.pub
