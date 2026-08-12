*** Settings ***
Documentation     Essential UoM Category resource suite. File order (NN_ prefixes)
...               encodes the mandated flow: CREATE (saves the category under test) ->
...               UPDATE -> GET -> EXISTS -> ARCHIVE -> SEARCH -> DELETE (cleanup,
...               always last). Teardown removes only fixtures created on the fly.
Resource          resources/essential.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Uom Seed Data
