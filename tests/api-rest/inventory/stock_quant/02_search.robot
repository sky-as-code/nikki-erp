*** Settings ***
Documentation     Searching Stock Balances. Phase 1 creates no movements, so the database
...               may legitimately hold no quants at all: these tests pin the search
...               contract and the projection rules, not a particular row count.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Inventory Org
Test Tags         inventory    stock_quant    search


*** Variables ***
${STOCK_QUANT_SCHEMA}    ${INVENTORY_SCHEMA_DIR}/stock_quant.json


*** Test Cases ***
Search Without Criteria Succeeds
    [Documentation]    An empty result is a valid answer here, so item_count is not asserted:
    ...    nothing in this phase produces a balance.
    ${resp}=    GET On Session    api    ${STOCK_QUANT_API}    params=${{ {'org_id': $INV_ORG_ID} }}
    Response Status Should Be    ${resp}    200
    Dictionary Should Contain Key    ${resp.json()}    items
    Dictionary Should Contain Key    ${resp.json()}    total

Search With Paging Succeeds
    ${resp}=    GET On Session    api    ${STOCK_QUANT_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'page': 0, 'size': 5} }}
    Response Status Should Be    ${resp}    200
    Should Be Equal As Integers    ${resp.json()}[size]    5
    Should Be Equal As Integers    ${resp.json()}[page]    0

Search Returns Available Quantity By Default
    [Documentation]    available_quantity is virtual, so it appears only because the derived
    ...    quant service fills it. Its absence from the default field set would mean the
    ...    service is not installed, which no other test would notice.
    ${resp}=    GET On Session    api    ${STOCK_QUANT_API}    params=${{ {'org_id': $INV_ORG_ID} }}
    Response Status Should Be    ${resp}    200
    Should Contain    ${resp.json()}[desired_fields]    available_quantity
    Should Contain    ${resp.json()}[desired_fields]    on_hand_quantity
    Should Contain    ${resp.json()}[desired_fields]    reserved_quantity

Search With Explicit Available Quantity Column Succeeds
    [Documentation]    A projection naming only the derived field still has to work: the
    ...    service adds the two operands it needs rather than returning a blank column.
    ${resp}=    GET On Session    api    ${STOCK_QUANT_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'fields': ['available_quantity']} }}
    Response Status Should Be    ${resp}    200

Search With Nonexist Column Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${STOCK_QUANT_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'fields': ['bla_bla_field']} }}    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field

Get With Not Found Id Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${STOCK_QUANT_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Get With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${STOCK_QUANT_API}/not-existing-1234567890123
    ...    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
