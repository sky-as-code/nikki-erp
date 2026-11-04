*** Settings ***
Documentation     Essential Sequence Definition resource suite (BR_CoreMart_Essential_Sequence §4,
...               §13). File order (NN_ prefixes): SCHEMA -> CREATE (saves the definition under
...               test) -> UPDATE -> PREVIEW -> GET/SEARCH -> CONFIG LOCK -> ARCHIVE ->
...               AUTHORIZATION. Definitions have no delete route (BR §4.4), so the teardown
...               archives the free definition and unwinds the Inventory fixture that issued one.
Resource          resources/sequence.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Sequence Fixtures
