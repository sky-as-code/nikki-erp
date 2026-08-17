*** Settings ***
Documentation     Warehouse resource suite. File order (NN_ prefixes) encodes the mandated
...               flow: CREATE (saves the warehouse under test, and with it the locations it
...               creates for itself) -> UPDATE -> GET -> EXISTS -> ARCHIVE -> SEARCH ->
...               DELETE, then the lifecycle and flow suites that need a live warehouse.
...               Teardown removes only fixtures created on the fly.
Resource          resources/inventory.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Inventory Seed Data
