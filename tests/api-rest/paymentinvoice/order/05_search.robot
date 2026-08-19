*** Settings ***
Documentation     Reading orders and transactions.
...
...               Neither is created by this suite — an order is minted by create_payment,
...               which needs a gateway — so these assert the endpoints answer correctly over
...               whatever the deployment holds, including nothing at all. The envelope is
...               checked directly rather than through Response Should Be Search Success,
...               which requires a non-empty result and would fail on a fresh database for a
...               reason that has nothing to do with the code under test.
Resource          resources/paymentinvoice.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Payment Invoice Org
Test Tags         paymentinvoice    order    search


*** Test Cases ***
Search Orders Succeeds
    ${resp}=    GET On Session    api    ${ORDER_API}    params=${{ {'org_id': $PAYINV_ORG_ID} }}
    Response Status Should Be    ${resp}    200
    ${body}=    Set Variable    ${resp.json()}
    Dictionary Should Contain Key    ${body}    items
    Dictionary Should Contain Key    ${body}    total
    Should Be Equal As Integers    ${body}[page]    0

Search Orders With Paging Succeeds
    ${resp}=    GET On Session    api    ${ORDER_API}
    ...    params=${{ {'org_id': $PAYINV_ORG_ID, 'page': 0, 'size': 5} }}
    Response Status Should Be    ${resp}    200
    Should Be Equal As Integers    ${resp.json()}[size]    5

Search Orders By Status Succeeds
    [Documentation]    The status filter is what the watchdog's index serves and what the
    ...    frontend's tabs use.
    ${resp}=    GET On Session    api    ${ORDER_API}
    ...    params=${{ {'org_id': $PAYINV_ORG_ID, 'graph': '{"if":["status", "=", "payment_success"]}'} }}
    Response Status Should Be    ${resp}    200
    Dictionary Should Contain Key    ${resp.json()}    items

Search Orders With Nonexist Field Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${ORDER_API}
    ...    params=${{ {'org_id': $PAYINV_ORG_ID, 'graph': '{"if":["bla_bla_field", "=", "x"]}'} }}
    ...    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field

Get Order With Not Found Id Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${ORDER_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Get Order With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${ORDER_API}/not-existing-1234567890123    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id

Search Transactions Succeeds
    ${resp}=    GET On Session    api    ${TRANSACTION_API}    params=${{ {'org_id': $PAYINV_ORG_ID} }}
    Response Status Should Be    ${resp}    200
    Dictionary Should Contain Key    ${resp.json()}    items

Get Transaction With Not Found Id Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${TRANSACTION_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}
