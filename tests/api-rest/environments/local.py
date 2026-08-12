"""Local environment for nikkierp API REST tests.

One Python variable file per environment (Bruno environment equivalent). Every value
can be overridden through process env so CI/CD can retarget without code changes.
Coremart ships its own file with the SAME variable names under
coremart/tests/api-rest/environments/.
"""
import os
from pathlib import Path

_PKI = Path(__file__).resolve().parents[3] / "scripts" / "cert" / "pki"


def _verify(default_ca):
    value = os.getenv("API_SSL_VERIFY", default_ca)
    if value in ("0", "false", "False", "no"):
        return False
    return value


API_HOST = os.getenv("API_HOST", "https://api.nikkierp.com:4433")

# Server certificate verification: path to a CA bundle, or False via API_SSL_VERIFY=false.
SSL_VERIFY = _verify(str(_PKI / "root-ca" / "nikkierp-root-ca.crt"))

# mTLS client certificate (bruno.json "clientCertificates" equivalent).
CLIENT_CERT = os.getenv("API_CLIENT_CERT", str(_PKI / "client-cert" / "active@nikki-erp.com.crt"))
CLIENT_KEY = os.getenv("API_CLIENT_KEY", str(_PKI / "client-cert" / "active@nikki-erp.com.key"))

# Sign-in flow. Route prefix is configurable because deployments may expose the flow
# under a different prefix (the Bruno collection used /v1/authn).
SIGNIN_API = os.getenv("API_SIGNIN_PREFIX", "/v1/iam/signin")
SIGNIN_USERNAME = os.getenv("API_TEST_USERNAME", "nguyen.van.an@nikki.com")
SIGNIN_PASSWORD = os.getenv("API_TEST_PASSWORD", "Passwo0rd123")

# Expected CORS preflight response headers (see CORE.HTTP.CORS_* in config.default.yaml).
CORS_ALLOW_METHODS = os.getenv("API_CORS_METHODS", "OPTIONS,HEAD,GET,POST,PATCH,DELETE")
CORS_ALLOW_HEADERS = os.getenv("API_CORS_HEADERS", "Accept,Authorization,Content-Type,Origin")

# Optional pre-existing seed entities. Left empty, the suites create a sample
# user/group on the fly and delete them in the suite teardown.
SEED_USER_ID = os.getenv("SEED_USER_ID", "")
SEED_GROUP_ID = os.getenv("SEED_GROUP_ID", "")
ORG_ID = os.getenv("ORG_ID", "")

# Credentials of an account holding ONLY the system `User` role, used by the Inventory
# permission regression test: Product master data grants no domain-wide read to that
# role, unlike Essential UoM, so such a user must be refused. Left empty the test skips
# rather than failing, because no environment can be assumed to have provisioned one.
PLAIN_USER_USERNAME = os.getenv("API_PLAIN_USER_USERNAME", "")
PLAIN_USER_PASSWORD = os.getenv("API_PLAIN_USER_PASSWORD", "")
