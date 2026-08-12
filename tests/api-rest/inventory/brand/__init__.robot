*** Settings ***
Documentation     Inventory Brand resource suite. File order (NN_ prefixes) encodes the
...               mandated flow: CREATE (saves the brand under test) -> UPDATE -> GET ->
...               EXISTS -> ARCHIVE -> SEARCH -> DELETE (cleanup, always last). Teardown
...               removes only fixtures created on the fly.
Resource          resources/inventory.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Inventory Seed Data
