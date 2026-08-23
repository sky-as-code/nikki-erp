*** Settings ***
Documentation     What makes settings more than a key-value table: the partial save, the reach
...               of each level, and the fan-out that lets a tenant enforce a value onto rows
...               that already exist. These behaviours span many rows of one transaction and
...               are invisible at the resource endpoints, so they are covered here.
Resource          resources/settings.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Settings Fixtures
Force Tags        settings    levels
