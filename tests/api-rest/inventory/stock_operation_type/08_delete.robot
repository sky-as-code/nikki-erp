*** Settings ***
Documentation     Deleting the Stock Operation Type under test — always the LAST suite,
...               doubling as cleanup. No transfer exists yet in this phase, so the type has
...               no dependant and deletes cleanly. Once transfers land, deleting a type a
...               transfer references must be refused in favour of archive (BR §4.2.1.5).
...               SEQ-FB-002: a type owns no sequence definition, so deleting one leaves the org's
...               inventory.transfer definition (and the other types' counters) as they were.
Resource          resources/inventory.resource
Resource          resources/sequence.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Stock Operation Type Under Test
Test Tags         inventory    stock_operation_type    delete


*** Test Cases ***
Delete Succeeds
    ${type}=    GET On Session    api    ${STOCK_OPERATION_TYPE_API}/${STOCK_OPERATION_TYPE_ID}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}
    Response Status Should Be    ${type}    200
    ${resp}=    DELETE On Session    api    ${STOCK_OPERATION_TYPE_API}/${STOCK_OPERATION_TYPE_ID}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}
    Response Should Be Delete Success    ${resp}    count=1


Delete Again With Same Id Fails
    [Documentation]    Idempotency check on the just-deleted id; clears the globals so the
    ...    type under test is not reused after this point.
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${STOCK_OPERATION_TYPE_API}/${STOCK_OPERATION_TYPE_ID}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}
    Set Global Variable    ${STOCK_OPERATION_TYPE_ID}    ${EMPTY}
    Set Global Variable    ${STOCK_OPERATION_TYPE_ETAG}    ${EMPTY}

Delete With Not Found Id Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${STOCK_OPERATION_TYPE_API}/${NOT_FOUND_ID}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Delete With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${STOCK_OPERATION_TYPE_API}/not-existing-1234567890123
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id

Delete Of A Type That Numbered A Transfer Leaves The Org Definition
    [Documentation]    SEQ-FB-002: the type's numbers came from its counter on the org's shared
    ...    inventory.transfer definition, which the org keeps numbering with: deleting the type
    ...    neither archives nor removes it. The transfer is deleted first: a live transfer would
    ...    hold the type itself through its foreign key.
    Ensure Inventory Location Under Test
    Ensure Stock Destination Location Under Test
    ${name}=    Unique Display Name    Robot Numbered Type
    ${resp}=    Post Stock Operation Type    ${{ {'name': {'en-US': $name}, 'operation_code': 'internal', 'reservation_method': 'manual', 'backorder_policy': 'never', 'org_id': $INV_ORG_ID} }}
    ${type_id}    ${type_etag}=    Response Should Be Create Success    ${resp}
    ${transfer_id}    ${transfer_etag}=    Create Stock Transfer    ${type_id}
    ${definition_id}=    Org Transfer Sequence Definition Id
    ${resp}=    DELETE On Session    api    ${STOCK_TRANSFER_API}/${transfer_id}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}
    Response Should Be Delete Success    ${resp}    count=1
    ${resp}=    DELETE On Session    api    ${STOCK_OPERATION_TYPE_API}/${type_id}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}
    Response Should Be Delete Success    ${resp}    count=1
    ${resp}=    GET On Session    api    ${SEQ_DEFINITION_API}/${definition_id}
    Response Status Should Be    ${resp}    200
    Should Not Be True    ${resp.json()}[item][is_archived]
    ...    msg=Deleting a type must leave the org's transfer definition in use
