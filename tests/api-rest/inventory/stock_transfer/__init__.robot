*** Settings ***
Documentation     Inventory Stock Transfer resource suite. File order (NN_ prefixes) encodes
...               the mandated flow: MODEL SCHEMA -> CREATE (saves the transfer under test) ->
...               UPDATE -> GET -> SEARCH -> LIFECYCLE (confirm/cancel and the state rules) ->
...               DELETE (cleanup, always last).
...
...               The end-to-end paths that actually move stock live in inventory/stock_flows,
...               because they span transfer, moves, move lines and balances on both sides and
...               belong to no single resource.
Resource          resources/inventory.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Inventory Seed Data
