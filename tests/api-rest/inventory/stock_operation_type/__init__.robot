*** Settings ***
Documentation     Inventory Stock Operation Type resource suite. File order (NN_ prefixes)
...               encodes the mandated flow: CREATE (saves the type under test) -> UPDATE ->
...               GET -> EXISTS -> ARCHIVE -> SEARCH -> DELETE (cleanup, always last).
...               Teardown removes only fixtures created on the fly.
Resource          resources/inventory.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Inventory Seed Data
