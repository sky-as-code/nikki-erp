*** Settings ***
Documentation     Settings module. The two resource suites cover the stored rows; the levels
...               suite covers what makes settings more than a key-value table — the partial
...               save, the per-level reach and the enforcement fan-out; the permissions suite
...               pins who may edit what.
Resource          resources/settings.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Settings Fixtures
Force Tags        settings
