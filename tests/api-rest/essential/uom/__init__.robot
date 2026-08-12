*** Settings ***
Documentation     Essential Unit of Measure resource suite. File order (NN_ prefixes)
...               encodes the mandated flow: CREATE (saves the unit under test) ->
...               UPDATE -> GET -> EXISTS -> ARCHIVE -> SEARCH -> CONVERT -> DELETE
...               (cleanup, always last). Teardown removes the category and reference
...               UoM the fixtures created on the fly.
Resource          resources/essential.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Uom Fixtures
