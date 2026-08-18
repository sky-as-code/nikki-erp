*** Settings ***
Documentation     The Purchase Agreement schema.
...
...               The one asymmetry worth pinning is archivability: an agreement IS archivable
...               where an order is not. An agreement is a standing arrangement that can fall out
...               of use without being cancelled, so archiving it takes it out of the working set
...               while leaving the orders drawn against it intact. An order has no such state —
...               its lifecycle is its status.
Resource          resources/purchase.resource
Suite Setup       Create Authorized API Session
Test Tags         purchase    purchase_agreement    schema


*** Test Cases ***
Get Purchase Agreement Model Schema
    ${resp}=    GET On Session    api    ${AGREEMENT_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Schema Declares The Core Fields
    ${resp}=    GET On Session    api    ${AGREEMENT_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    FOR    ${name}    IN    code    reference    agreement_type    status    vendor_id
    ...    buyer_id    currency_id    start_date    end_date    org_id
        Dictionary Should Contain Key    ${fields}    ${name}
    END

Schema Declares The Agreement Status Enum
    [Documentation]    Four statuses, and closed and cancelled mean different things: a closed
    ...    agreement ran its course and the orders drawn against it stand, where a cancelled one
    ...    was called off.
    ${resp}=    GET On Session    api    ${AGREEMENT_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${values}=    Set Variable    ${resp.json()}[fields][status][data_type][options][enumValues]
    FOR    ${status}    IN    draft    confirmed    closed    cancelled
        Should Contain    ${values}    ${status}
    END
    Length Should Be    ${values}    4

Schema Declares The Agreement Type Enum
    [Documentation]    A blanket order commits to quantities at agreed prices and tracks what
    ...    has been drawn against it; a purchase template is a reusable skeleton with no
    ...    commitment attached.
    ${resp}=    GET On Session    api    ${AGREEMENT_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${values}=    Set Variable    ${resp.json()}[fields][agreement_type][data_type][options][enumValues]
    Should Contain    ${values}    blanket_order
    Should Contain    ${values}    purchase_template
    Length Should Be    ${values}    2

Schema Declares An Archived Field
    [Documentation]    Unlike the order. See this suite's documentation for why the two differ.
    ${resp}=    GET On Session    api    ${AGREEMENT_API}/meta/schema
    Response Status Should Be    ${resp}    200
    Dictionary Should Contain Key    ${resp.json()}[fields]    is_archived

Agreement Line Schema Stores No Ordered Quantity
    [Documentation]    §41. ordered_quantity is DERIVED on read, never stored: it changes
    ...    whenever any referencing order is confirmed or cancelled, so a stored copy would need
    ...    invalidating from the order side — and the day one path forgot, the agreement would
    ...    misreport its own drawdown with nothing to reconcile against.
    ${resp}=    GET On Session    api    ${AGREEMENT_LINE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    Dictionary Should Not Contain Key    ${resp.json()}[fields]    ordered_quantity
