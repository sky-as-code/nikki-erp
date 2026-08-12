*** Settings ***
Documentation     Inventory Stock Quant (stock balance) suite. Unlike every other Inventory
...               resource this one has no create/update/delete lifecycle: a balance is the
...               running total of completed movements, not a document a client writes. The
...               files therefore read SCHEMA -> SEARCH -> REJECT WRITES -> AVAILABLE
...               QUANTITY -> PERMISSIONS rather than the usual CRUD order.
Resource          resources/inventory.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Inventory Seed Data
