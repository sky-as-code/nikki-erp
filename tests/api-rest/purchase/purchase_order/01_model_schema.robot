*** Settings ***
Documentation     The Purchase Order schema, served by the dynamic resource engine.
...
...               This suite is where the module's two most load-bearing schema decisions are
...               pinned: one resource carries the whole RFQ-to-PO cycle (PUR-R1), and is_locked
...               is a boolean rather than a status (PUR-R2). Both are the kind of thing a later
...               "tidy-up" would undo without realising what it cost.
Resource          resources/purchase.resource
Suite Setup       Create Authorized API Session
Test Tags         purchase    purchase_order    schema


*** Test Cases ***
Get Purchase Order Model Schema
    ${resp}=    GET On Session    api    ${PURCHASE_ORDER_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Schema Declares The Core Fields
    ${resp}=    GET On Session    api    ${PURCHASE_ORDER_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    FOR    ${name}    IN    code    status    vendor_id    buyer_id    currency_id
    ...    order_deadline    expected_arrival    priority    is_locked    vendor_acknowledged
    ...    untaxed_amount    tax_amount    total_amount    approval_required    org_id
        Dictionary Should Contain Key    ${fields}    ${name}
    END

Schema Carries The Whole RFQ To PO Cycle In One Status Field
    [Documentation]    PUR-R1. An RFQ and the purchase order it becomes are ONE record:
    ...    confirming changes the status and nothing about its identity. Five statuses, and the
    ...    two RFQ ones are among them — a separate request_for_quote resource would mean a
    ...    vendor's quotation reference stopped matching the order it turned into.
    ${resp}=    GET On Session    api    ${PURCHASE_ORDER_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${values}=    Set Variable    ${resp.json()}[fields][status][data_type][options][enumValues]
    FOR    ${status}    IN    rfq    rfq_sent    to_approve    purchase_order    cancelled
        Should Contain    ${values}    ${status}
    END
    Length Should Be    ${values}    5

Schema Declares Locked As A Boolean And Never A Status
    [Documentation]    PUR-R2. An order is locked AND confirmed, not locked INSTEAD OF
    ...    confirmed. Making lock a status would force the transition table to route around it
    ...    and would lose the confirmation state the moment somebody locked the document.
    ${resp}=    GET On Session    api    ${PURCHASE_ORDER_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    Should Be Equal    ${fields}[is_locked][data_type][name]    boolean
    Should Not Contain    ${fields}[status][data_type][options][enumValues]    locked

Schema Declares No Archived Field
    [Documentation]    PUR-R2. An order is not archivable at all: its lifecycle IS its status,
    ...    and an archived-but-open order would be a document withdrawn from view while still
    ...    committing the business to a purchase. The agreement, which is archivable, does carry
    ...    one — see the agreement schema suite.
    ${resp}=    GET On Session    api    ${PURCHASE_ORDER_API}/meta/schema
    Response Status Should Be    ${resp}    200
    Dictionary Should Not Contain Key    ${resp.json()}[fields]    is_archived

Schema Marks The Server Owned Fields As Not Updatable
    [Documentation]    §10.2 and §55.12: the totals are computed from the lines and a client
    ...    value for them is ignored, not trusted. The status and the approval evidence are the
    ...    same — changing them is what the lifecycle actions are for, and an update that could
    ...    set them would let a role with `update` confirm and self-approve an order.
    ${resp}=    GET On Session    api    ${PURCHASE_ORDER_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    FOR    ${name}    IN    code    status    confirmed_at    is_locked    vendor_acknowledged
    ...    untaxed_amount    tax_amount    total_amount    approval_required    approved_by
    ...    approved_at
        Should Be True    ${fields}[${name}][no_update]
        ...    msg=${name} must be no_update; a client must not be able to write it directly
    END

Schema Declares A Record Label Field
    [Documentation]    Without record_label_field a relation picker falls back to the raw ULID,
    ...    so an order shows as an opaque id wherever another resource points at it.
    ${resp}=    GET On Session    api    ${PURCHASE_ORDER_API}/meta/schema
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[record_label_field]    code

Schema Declares The Priority Enum
    ${resp}=    GET On Session    api    ${PURCHASE_ORDER_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${values}=    Set Variable    ${resp.json()}[fields][priority][data_type][options][enumValues]
    Should Contain    ${values}    normal
    Should Contain    ${values}    urgent
    Length Should Be    ${values}    2

Schema Holds Cross Module References As Plain Ids
    [Documentation]    D5a. vendor_id, currency_id and agreement_id are plain ulid fields with
    ...    no edge: a foreign key across a module boundary would make Purchase's schema depend
    ...    on another module's tables, and `-createsql -module=purchase` emits only Purchase's,
    ...    so the Atlas diff would fail. Validation is by port call instead.
    ${resp}=    GET On Session    api    ${PURCHASE_ORDER_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    FOR    ${name}    IN    vendor_id    buyer_id    currency_id
        Should Be Equal    ${fields}[${name}][data_type][name]    ulid
    END
