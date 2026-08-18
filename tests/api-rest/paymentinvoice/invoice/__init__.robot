*** Settings ***
Documentation     Payment & Invoice invoice suite. File order (NN_ prefixes) encodes the
...               mandated flow: CREATE (saves the invoice under test) -> UPDATE -> GET ->
...               EXISTS -> ISSUE -> SEARCH -> DELETE (cleanup, always last).
...
...               ISSUE takes the archive slot the other resources use. An invoice has a real
...               lifecycle of its own — draft becomes issued and an issued invoice is an
...               accounting document — so closing one is the transition worth covering, and
...               it is placed after the read suites because it is irreversible.
...               Teardown removes only fixtures created on the fly.
Resource          resources/paymentinvoice.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Payment Invoice Seed Data
