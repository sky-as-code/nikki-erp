*** Settings ***
Documentation     Creating Stock Operation Types. The first test saves the type under test
...               (${STOCK_OPERATION_TYPE_ID}/${STOCK_OPERATION_TYPE_ETAG}) consumed by the
...               later suites and deleted last by 08_delete.robot.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Inventory Location Under Test
Test Tags         inventory    stock_operation_type    create


*** Test Cases ***
Create With Required Fields Succeeds
    ${name}=    Unique Display Name    Robot Receipt Type
    ${code}=    Unique Code    stkop
    ${resp}=    POST On Session    api    ${STOCK_OPERATION_TYPE_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'operation_code': 'incoming', 'reservation_method': 'at_confirmation', 'backorder_policy': 'ask', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${STOCK_OPERATION_TYPE_ID}    ${id}
    Set Global Variable    ${STOCK_OPERATION_TYPE_ETAG}    ${etag}
    Set Global Variable    ${STOCK_OPERATION_TYPE_CODE}    ${code}

Create With Default Locations Succeeds
    [Documentation]    The default source and destination are what a transfer of this type
    ...    pre-fills, so both must accept a real location reference.
    ${name}=    Unique Display Name    Robot Located Type
    ${code}=    Unique Code    locop
    ${resp}=    POST On Session    api    ${STOCK_OPERATION_TYPE_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'operation_code': 'internal', 'reservation_method': 'manual', 'backorder_policy': 'never', 'default_source_location_id': $INVENTORY_LOCATION_ID, 'default_destination_location_id': $INVENTORY_LOCATION_ID, 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${resp}=    GET On Session    api    ${STOCK_OPERATION_TYPE_API}/${id}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/stock_operation_type.json    200
    Should Be Equal    ${item}[default_source_location_id]    ${INVENTORY_LOCATION_ID}
    DELETE On Session    api    ${STOCK_OPERATION_TYPE_API}/${id}    expected_status=any

Create With Unknown Operation Code Fails
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Bad Operation
    ${code}=    Unique Code    badop
    ${resp}=    POST On Session    api    ${STOCK_OPERATION_TYPE_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'operation_code': 'sideways', 'reservation_method': 'manual', 'backorder_policy': 'ask', 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    201
    ...    msg=An operation_code outside the declared enum must not be accepted

Create With Unknown Reservation Method Fails
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Bad Reservation
    ${code}=    Unique Code    badres
    ${resp}=    POST On Session    api    ${STOCK_OPERATION_TYPE_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'operation_code': 'incoming', 'reservation_method': 'whenever', 'backorder_policy': 'ask', 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    201
    ...    msg=A reservation_method outside the declared enum must not be accepted

Create With Duplicate Code And Org Fails
    [Documentation]    composite_uniques on stock_operation_type.json is ["code", "org_id"].
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Duplicate Operation
    ${resp}=    POST On Session    api    ${STOCK_OPERATION_TYPE_API}
    ...    json=${{ {'code': $STOCK_OPERATION_TYPE_CODE, 'name': {'en-US': $name}, 'operation_code': 'incoming', 'reservation_method': 'manual', 'backorder_policy': 'ask', 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Response Should Be Duplicate Values Error    ${resp}    code    org_id

Create With Missing Required Fields Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${STOCK_OPERATION_TYPE_API}    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    code    name    operation_code    org_id
