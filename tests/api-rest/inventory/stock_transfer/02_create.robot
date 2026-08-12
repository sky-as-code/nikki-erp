*** Settings ***
Documentation     Creating Stock Transfers (BR §4.2.3.4, AC-STOCK-003). The first test saves the
...               transfer under test (${STOCK_TRANSFER_ID}/${STOCK_TRANSFER_ETAG}) consumed by
...               the later suites and deleted last by 08_delete.robot.
...
...               Creating a transfer takes no stock: it records a demand, and nothing about a
...               balance changes until the transfer is reserved and then validated.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Stock Location Under Test
...               AND    Ensure Stock Destination Location Under Test
...               AND    Ensure Internal Operation Type Under Test
Test Tags         inventory    stock_transfer    create


*** Test Cases ***
Create With Required Fields Succeeds
    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}
    ...    json=${{ {'operation_type_id': $INTERNAL_OPERATION_TYPE_ID, 'source_location_id': $STOCK_LOCATION_ID, 'destination_location_id': $STOCK_DEST_LOCATION_ID, 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${STOCK_TRANSFER_ID}    ${id}
    Set Global Variable    ${STOCK_TRANSFER_ETAG}    ${etag}

Create Generates The Transfer Number
    [Documentation]    The number identifies the document on paperwork and in other modules, so
    ...    it is generated rather than client-supplied: two clients picking the same one would
    ...    collide, and one could otherwise impersonate another's reference.
    ${resp}=    GET On Session    api    ${STOCK_TRANSFER_API}/${STOCK_TRANSFER_ID}
    Response Status Should Be    ${resp}    200
    ${item}=    Set Variable    ${resp.json()}[data]
    Should Not Be Empty    ${item}[transfer_number]
    ...    msg=Create must generate a transfer number

Create Starts In Draft
    [Documentation]    AC-STOCK-003. A transfer that could be created in any other state would be
    ...    a movement with nothing behind it.
    ${resp}=    GET On Session    api    ${STOCK_TRANSFER_API}/${STOCK_TRANSFER_ID}
    ${item}=    Set Variable    ${resp.json()}[data]
    Should Be Equal    ${item}[status]    draft

Create Snapshots The Operation Type Policies
    [Documentation]    BR §4.2.1.4 and §4.2.3.4. The fixture type is internal/manual/always/
    ...    partial, and the transfer must carry its own copy of each: reconfiguring the type
    ...    afterwards must not reinterpret a transfer already created.
    ${resp}=    GET On Session    api    ${STOCK_TRANSFER_API}/${STOCK_TRANSFER_ID}
    ${item}=    Set Variable    ${resp.json()}[data]
    Should Be Equal    ${item}[operation_code]    internal
    Should Be Equal    ${item}[reservation_method]    manual
    Should Be Equal    ${item}[backorder_policy]    always
    Should Be Equal    ${item}[shipping_policy]    partial

Create Overrides A Client Supplied Status
    [Documentation]    A client echoing a record back should not fail for carrying a status, but
    ...    must not be able to create a transfer that claims to be done either. The server-owned
    ...    field is stamped over rather than rejected.
    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}
    ...    json=${{ {'operation_type_id': $INTERNAL_OPERATION_TYPE_ID, 'source_location_id': $STOCK_LOCATION_ID, 'destination_location_id': $STOCK_DEST_LOCATION_ID, 'status': 'done', 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    IF    ${resp.status_code} == 201
        ${id}=    Set Variable    ${resp.json()}[data][id]
        ${check}=    Get On Session    api    ${STOCK_TRANSFER_API}/${id}
        Should Be Equal    ${check.json()}[data][status]    draft
        ...    msg=A client-supplied status must not survive create
        DELETE On Session    api    ${STOCK_TRANSFER_API}/${id}    expected_status=any
    END

Create With The Same Source And Destination Fails
    [Documentation]    BR §4.2.3.4. Such a transfer moves nothing, yet would still generate moves
    ...    and consume reservations against the balance it is about to put straight back.
    [Tags]    negative
    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}
    ...    json=${{ {'operation_type_id': $INTERNAL_OPERATION_TYPE_ID, 'source_location_id': $STOCK_LOCATION_ID, 'destination_location_id': $STOCK_LOCATION_ID, 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    201
    ...    msg=A transfer's source and destination must differ

Create With An Unknown Operation Type Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}
    ...    json=${{ {'operation_type_id': $NOT_FOUND_ID, 'source_location_id': $STOCK_LOCATION_ID, 'destination_location_id': $STOCK_DEST_LOCATION_ID, 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    201
    ...    msg=A transfer must reference an operation type that exists

Create With An Archived Operation Type Fails
    [Documentation]    BR §4.2.1.5, AC-STOCK-030. Archiving withdraws a type from new business.
    ...    The check is at create time, not read time, because a transfer created before the type
    ...    was archived must still resolve it (AC-STOCK-031).
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Archived Type
    ${code}=    Unique Code    arcop
    ${resp}=    POST On Session    api    ${STOCK_OPERATION_TYPE_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'operation_code': 'internal', 'reservation_method': 'manual', 'backorder_policy': 'ask', 'org_id': $INV_ORG_ID} }}
    ${type_id}    ${type_etag}=    Response Should Be Create Success    ${resp}

    POST On Session    api    ${STOCK_OPERATION_TYPE_API}/${type_id}/archived
    ...    json=${{ {'is_archived': True, 'etag': $type_etag} }}    expected_status=any

    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}
    ...    json=${{ {'operation_type_id': $type_id, 'source_location_id': $STOCK_LOCATION_ID, 'destination_location_id': $STOCK_DEST_LOCATION_ID, 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    201
    ...    msg=An archived operation type must not start a new transfer

Create With Missing Required Fields Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}    json=${{ {} }}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    201
    ...    msg=A transfer needs an operation type, both locations and an org
