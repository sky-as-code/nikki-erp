*** Settings ***
Documentation     Essential Sequence Counter resource suite: read-only (BR §8). Counters are
...               written only by NextValue, which a consumer reaches through Essential's port, so
...               the fixture issues a number by creating a stock transfer.
Resource          resources/sequence.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Sequence Fixtures
