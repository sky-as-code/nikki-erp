*** Settings ***
Documentation     Inventory Product Variant resource suite. File order (NN_ prefixes)
...               encodes the mandated flow: CREATE (saves the variant under test) ->
...               UPDATE -> GET -> EXISTS -> ARCHIVE -> SEARCH -> DELETE (cleanup, always
...               last). The variant is the concrete transactable SKU, so the suite pins
...               the combination uniqueness of BR-PROD-VAR-002 and the inheritance rules
...               that keep the template the single source of truth.
Resource          resources/inventory.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Inventory Seed Data
