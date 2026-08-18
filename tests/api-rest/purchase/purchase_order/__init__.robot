*** Settings ***
Documentation     Purchase Order resource suite. File order (NN_ prefixes) encodes the mandated
...               flow: CREATE (saves the order under test) -> UPDATE -> GET -> EXISTS ->
...               SEARCH -> the lifecycle actions -> DELETE (cleanup, always last).
...
...               DELETE is last because an order is deletable only from cancelled (BR 24), so
...               it has to follow the cancel the lifecycle suite performs. Running it earlier
...               would strand the order and every suite after it.
Resource          resources/purchase.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Purchase Seed Data
