*** Settings ***
Documentation     Essential Language suite: public_use for the org-less reads, and the per-org
...               activation guarded by the active-set rules.
Resource          resources/essential.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Org Under Test
Suite Teardown    Delete Language Fixtures
