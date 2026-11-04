*** Settings ***
Documentation     SAL-SEQ-001. Every order gets an order number from the org's `sales.order`
...               Sequence (SO + 5-digit number, per org) and a public reference from
...               `sales.order.public_reference` (CORE + 6 Crockford Base32 characters, FF1
...               pseudorandom). The public reference is what the kiosk shows the customer, so it
...               must be unique, unguessable in order and never change once issued.
Resource          resources/sales.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Sales Order Inputs
Suite Teardown    Delete Numbering Orders
Test Tags         sales    sales_order    sequence


*** Test Cases ***
Create Order Issues An SO Number
    ${order}=    Create Sales Order
    Remember Order    ${order}
    Should Match Regexp    ${order}[order_number]    ^SO\\d{5,}$
    Set Suite Variable    ${FIRST_ORDER}    ${order}

Create Order Issues A CORE Public Reference
    Should Match Regexp    ${FIRST_ORDER}[public_reference]    ^CORE${SALES_CROCKFORD_CHAR}{6}$

The Stored Order Carries The Issued Numbers
    ${item}=    Read Sales Order    ${FIRST_ORDER}[sales_order_id]
    Should Be Equal    ${item}[order_number]    ${FIRST_ORDER}[order_number]
    Should Be Equal    ${item}[public_reference]    ${FIRST_ORDER}[public_reference]

The Next Order Takes A Higher Number And Another Reference
    ${second}=    Create Sales Order
    Remember Order    ${second}
    ${first_number}=    Evaluate    int($FIRST_ORDER['order_number'][2:])
    ${second_number}=    Evaluate    int($second['order_number'][2:])
    Should Be True    ${second_number} > ${first_number}
    Should Not Be Equal    ${second}[public_reference]    ${FIRST_ORDER}[public_reference]
    Should Match Regexp    ${second}[public_reference]    ^CORE${SALES_CROCKFORD_CHAR}{6}$

Public References Are Unique Across Many Orders
    [Documentation]    FF1 is a permutation of the counter, so distinct numbers can never
    ...    encrypt to the same reference.
    ${references}=    Create List
    FOR    ${index}    IN RANGE    10
        ${order}=    Create Sales Order
        Remember Order    ${order}
        Append To List    ${references}    ${order}[public_reference]
    END
    List Should Not Contain Duplicates    ${references}

The Numbers Cannot Be Changed By An Update
    [Documentation]    order_number and public_reference are no_update: printed receipts and the
    ...    kiosk screen quote them.
    ${item}=    Read Sales Order    ${FIRST_ORDER}[sales_order_id]
    PATCH On Session    api    ${SALES_ORDER_API}/${FIRST_ORDER}[sales_order_id]
    ...    json=${{ {'org_id': $SALES_ORG_ID, 'etag': $item['etag'], 'order_number': 'SO-HIJACK', 'public_reference': 'COREXXXXXX'} }}
    ...    expected_status=any
    ${after}=    Read Sales Order    ${FIRST_ORDER}[sales_order_id]
    Should Be Equal    ${after}[order_number]    ${FIRST_ORDER}[order_number]
    Should Be Equal    ${after}[public_reference]    ${FIRST_ORDER}[public_reference]

The Order Is Found By Its Public Reference
    ${graph}=    Evaluate    __import__('json').dumps({'if': ['public_reference', '=', $FIRST_ORDER['public_reference']]})
    ${resp}=    GET On Session    api    ${SALES_ORDER_API}
    ...    params=${{ {'org_id': $SALES_ORG_ID, 'graph': $graph, 'fields': 'id,public_reference'} }}
    Response Status Should Be    ${resp}    200
    Length Should Be    ${resp.json()}[items]    1
    Should Be Equal    ${resp.json()}[items][0][id]    ${FIRST_ORDER}[sales_order_id]

A Replayed Create Returns The Same Numbers
    [Documentation]    create_order is idempotent by idempotency key: the replay is the same order,
    ...    so it must not burn a second number or reference.
    ${suffix}=    Unique Suffix
    ${first}=    Create Sales Order    idempotency_key=robot-seq-${suffix}
    Remember Order    ${first}
    ${replay}=    Create Sales Order    idempotency_key=robot-seq-${suffix}
    Should Be True    ${replay}[already_existed]
    Should Be Equal    ${replay}[sales_order_id]    ${first}[sales_order_id]
    Should Be Equal    ${replay}[order_number]    ${first}[order_number]
    Should Be Equal    ${replay}[public_reference]    ${first}[public_reference]


*** Keywords ***
Remember Order
    [Arguments]    ${order}
    ${ids}=    Get Variable Value    ${NUMBERING_ORDER_IDS}    ${None}
    IF    $ids is None
        ${ids}=    Create List
        Set Suite Variable    ${NUMBERING_ORDER_IDS}    ${ids}
    END
    Append To List    ${NUMBERING_ORDER_IDS}    ${order}[sales_order_id]

Delete Numbering Orders
    ${ids}=    Get Variable Value    ${NUMBERING_ORDER_IDS}    ${None}
    IF    $ids is None    RETURN
    FOR    ${id}    IN    @{ids}
        Delete Sales Order Fixture    ${id}
    END
