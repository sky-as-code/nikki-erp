*** Settings ***
Documentation     Contacts Relationship resource suite. File order (NN_ prefixes) encodes the
...               mandated flow: CREATE (saves the relationship under test) -> UPDATE -> GET
...               -> EXISTS -> ARCHIVE -> SEARCH -> DELETE (cleanup, always last). Teardown
...               removes only fixtures created on the fly.
Resource          resources/contacts.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Contacts Seed Data
