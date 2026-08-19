*** Settings ***
Documentation     Contacts Vendor Profile resource suite. File order (NN_ prefixes) encodes the
...               mandated flow: CREATE (saves the profile under test) -> UPDATE -> GET ->
...               EXISTS -> ARCHIVE -> SEARCH -> DELETE (cleanup, always last). Teardown
...               removes only fixtures created on the fly.
...
...               The vendor profile is a sidecar on a party rather than a resource of its own
...               standing: "this party is a vendor" is expressed as "this party has a profile
...               row", which is what makes it checkable. Purchase reads it through a port to
...               validate a purchase order's vendor_id and to default the order's currency,
...               payment terms and expected arrival.
Resource          resources/contacts.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Contacts Seed Data
