*** Settings ***
Documentation     Vendor Product Price: what each supplier offers a product at, by quantity, unit
...               and validity (sections 21-30).
...
...               The resource is master data rather than a document. It has no lifecycle and no
...               status: an offer is recorded, corrected, and eventually archived, and none of
...               those is a state machine. That is why this suite has no lifecycle file where the
...               order and agreement suites do.
Resource          resources/purchase.resource
Suite Setup       Create Authorized API Session
