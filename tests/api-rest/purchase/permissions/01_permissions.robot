*** Settings ***
Documentation     What the Purchase module exposes to a caller, and what it deliberately does
...               not. This suite pins [PUR-013]'s no-blanket-grant decision.
...
...               Unit of Measure grants the system "User" role a domain-wide read, so any user
...               can pick a unit while filling in an unrelated form. Purchase data is not
...               comparable: an order carries what the business pays, to whom, on what terms,
...               and the audit trail carries who approved it. There are NO iam_entitlements
...               rows for purchase, so access follows explicitly assigned roles — the same
...               choice Products and Contacts made.
...
...               Without a test, a blanket grant added later would expose every order, price and
...               approval in the system and nothing in the tree would notice.
Resource          resources/purchase.resource
Suite Setup       Create Authorized API Session
Test Tags         purchase    permissions


*** Variables ***
${AUDIT_SCHEMA}    ${PURCHASE_SCHEMA_DIR}/purchase_audit_event.json


*** Test Cases ***
Purchase Seeds No Domain Wide Entitlement
    [Documentation]    The check that matters. An entitlement granting purchase access to
    ...    everybody would make every other permission in this module decorative.
    ${resp}=    GET On Session    api    /v1/iam/iam_entitlement    params=${{ {'size': 200} }}
    Response Status Should Be    ${resp}    200
    FOR    ${item}    IN    @{resp.json()}[items]
        ${action}=    Get From Dictionary    ${item}    action_name    ${EMPTY}
        ${resource}=    Get From Dictionary    ${item}    resource_name    ${EMPTY}
        Should Not Start With    ${resource}    purchase_
        ...    msg=Purchase must not seed a domain-wide entitlement; found one for ${resource}.${action}
    END

Every Purchase Resource Is Reachable
    [Documentation]    Seven schemas, seven engines, seven route sets. A schema registered
    ...    without an engine has no HTTP surface at all, and one whose route drifted from its
    ...    schema name is refused rather than 404'd — which reads as a permission problem and
    ...    sends the reader looking in the wrong place.
    FOR    ${api}    IN    ${PURCHASE_ORDER_API}    ${PURCHASE_ORDER_LINE_API}
    ...    ${AGREEMENT_API}    ${AGREEMENT_LINE_API}    ${PURCHASE_CONFIG_API}
    ...    ${SOURCING_GROUP_API}    ${AUDIT_EVENT_API}
        ${resp}=    GET On Session    api    ${api}/meta/schema
        Response Status Should Be    ${resp}    200
    END

The Audit Trail Refuses Client Writes
    [Documentation]    PUR-R6. An audit event is written by the system inside the same
    ...    transaction as the transition it records, and by nothing else. A client-written event
    ...    would be a claim that something happened, sitting in the same table as the events that
    ...    did, with no way for a reader to tell them apart — which destroys the value of the
    ...    trail rather than adding to it.
    ...
    ...    The action is REFUSED rather than removed, so a caller gets a 400 naming the reason
    ...    instead of a 404 that reads as a wrong URL.
    Ensure Purchase Org
    ${resp}=    POST On Session    api    ${AUDIT_EVENT_API}    json=${{ {'entity_type': 'purchase_order', 'entity_id': '01JZZZZZZZZZZZZZZZZZZZZZZZ', 'action': 'confirm', 'org_id': $PURCHASE_ORG_ID} }}    expected_status=any
    Response Should Be Purchase Violation    ${resp}    purchase_audit_event.not_client_writable

The Sourcing Group Refuses Direct Creation
    [Documentation]    §28. The group is created by adding an alternative to an order and reaped
    ...    when fewer than two remain, so a hand-made one would be an empty container that
    ...    nothing reaps.
    Ensure Purchase Org
    ${resp}=    POST On Session    api    ${SOURCING_GROUP_API}    json=${{ {'org_id': $PURCHASE_ORG_ID} }}    expected_status=any
    Response Should Be Purchase Violation    ${resp}    purchase_sourcing_group.not_client_writable

The Audit Trail Is Readable
    [Documentation]    Read is the one verb the audit event does have: the trail exists to be
    ...    consulted, and seeding it write-only would make it useless.
    ${resp}=    GET On Session    api    ${AUDIT_EVENT_API}    params=${{ {'size': 5} }}
    Response Status Should Be    ${resp}    200
    Dictionary Should Contain Key    ${resp.json()}    items
