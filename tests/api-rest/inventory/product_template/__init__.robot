*** Settings ***
Documentation     Inventory Product Template resource suite. File order (NN_ prefixes)
...               encodes the mandated flow: CREATE (saves the template under test) ->
...               UPDATE -> GET -> EXISTS -> ARCHIVE -> SEARCH -> DELETE (cleanup, always
...               last). The template is the catalog-level half of the Odoo split, so the
...               suite pins what it owns and — just as importantly — what it must not.
Resource          resources/inventory.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Inventory Seed Data
