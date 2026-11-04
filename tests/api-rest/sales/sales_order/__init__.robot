*** Settings ***
Documentation     Sales Order resource suite. For now it covers document numbering only
...               (SAL-SEQ-001): the order number and the public reference both come from the
...               org's Sequence definitions. Orders it raises are drafts and are deleted again.
Resource          resources/sales.resource
Suite Setup       Create Authorized API Session
