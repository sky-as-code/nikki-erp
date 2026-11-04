*** Settings ***
Documentation     Essential Sequence Issuance resource suite: read-only (BR §6.3, §8). An
...               issuance is the record of one number handed out; the fixture issues one by
...               creating a stock transfer.
Resource          resources/sequence.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Sequence Fixtures
