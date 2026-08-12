*** Settings ***
Documentation     End-to-end stock movement paths. These are the tests the whole movement engine
...               exists to pass: they span transfer, moves, move lines and the balances on both
...               sides, so they belong to no single resource suite.
...
...               Every flow here starts by putting stock somewhere with a receipt, because
...               Phase 1 made the quant read-only to clients: there is no way to seed a balance
...               except by validating an incoming transfer, which is exactly the point.
Resource          resources/inventory.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Inventory Seed Data
