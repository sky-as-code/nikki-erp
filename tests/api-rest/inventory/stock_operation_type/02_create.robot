*** Settings ***
Documentation     Creating Stock Operation Types. The first test saves the type under test
...               (${STOCK_OPERATION_TYPE_ID}/${STOCK_OPERATION_TYPE_ETAG}) consumed by the
...               later suites and deleted last by 08_delete.robot.
Resource          resources/inventory.resource
Resource          resources/sequence.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Inventory Location Under Test
Test Tags         inventory    stock_operation_type    create


*** Keywords ***
Transfer Number Of
    [Arguments]    ${transfer_id}
    ${resp}=    GET On Session    api    ${STOCK_TRANSFER_API}/${transfer_id}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}
    Response Status Should Be    ${resp}    200
    RETURN    ${resp.json()}[item][transfer_number]


*** Test Cases ***
Create With Required Fields Succeeds
    ${name}=    Unique Display Name    Robot Receipt Type
    ${resp}=    Post Stock Operation Type    ${{ {'name': {'en-US': $name}, 'operation_code': 'incoming', 'reservation_method': 'at_confirmation', 'backorder_policy': 'ask', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${STOCK_OPERATION_TYPE_ID}    ${id}
    Set Global Variable    ${STOCK_OPERATION_TYPE_ETAG}    ${etag}

Create Registers No Sequence Definition
    [Documentation]    SEQ-FB-002: definitions are registered only by the org lifecycle or the
    ...    seeds. Every type numbers its transfers from the org's one inventory.transfer
    ...    definition, consumer-formatted ({warehouse}/{prefix}/{n} is Inventory's), on its own counter.
    ${resp}=    GET On Session    api    ${STOCK_OPERATION_TYPE_API}/${STOCK_OPERATION_TYPE_ID}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/stock_operation_type.json    200
    Should Be Equal    ${item}[warehouse_id]    ${WAREHOUSE_ID}
    Dictionary Should Not Contain Key    ${item}    sequence_definition_id
    ${definition_id}=    Org Transfer Sequence Definition Id
    ${definition}=    GET On Session    api    ${SEQ_DEFINITION_API}/${definition_id}
    Response Status Should Be    ${definition}    200
    ${def}=    Set Variable    ${definition.json()}[item]
    Should Be Equal    ${def}[owner_module]    inventory
    Should Be Equal    ${def}[configuration_scope]    org
    Should Be Equal    ${def}[format_owner]    consumer
    Should Be Equal    ${def}[scope_type]    org.inventory_optype
    Dictionary Should Not Contain Key    ${def}    algorithm_params

Two Types Of One Org Number Their Transfers Independently
    [Documentation]    SEQ-FB-002: one definition, one counter per type, so each fresh type's first
    ...    transfer is {warehouse}/{its prefix}/00001.
    Ensure Stock Destination Location Under Test
    ${in_name}=    Unique Display Name    Robot Counter In
    ${out_name}=    Unique Display Name    Robot Counter Out
    ${resp}=    Post Stock Operation Type    ${{ {'name': {'en-US': $in_name}, 'operation_code': 'internal', 'reservation_method': 'manual', 'backorder_policy': 'never', 'org_id': $INV_ORG_ID} }}
    ${in_type}    ${in_etag}=    Response Should Be Create Success    ${resp}
    ${resp}=    Post Stock Operation Type    ${{ {'name': {'en-US': $out_name}, 'operation_code': 'internal', 'reservation_method': 'manual', 'backorder_policy': 'never', 'org_id': $INV_ORG_ID} }}
    ${out_type}    ${out_etag}=    Response Should Be Create Success    ${resp}
    ${in_transfer}    ${etag}=    Create Stock Transfer    ${in_type}
    ${out_transfer}    ${etag}=    Create Stock Transfer    ${out_type}
    ${in_number}=    Transfer Number Of    ${in_transfer}
    ${out_number}=    Transfer Number Of    ${out_transfer}
    Should End With    ${in_number}    /00001
    Should End With    ${out_number}    /00001
    Should Not Be Equal    ${in_number}    ${out_number}
    [Teardown]    Run Keywords
    ...    DELETE On Session    api    ${STOCK_TRANSFER_API}/${in_transfer}    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    ...    AND    DELETE On Session    api    ${STOCK_TRANSFER_API}/${out_transfer}    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    ...    AND    DELETE On Session    api    ${STOCK_OPERATION_TYPE_API}/${in_type}    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    ...    AND    DELETE On Session    api    ${STOCK_OPERATION_TYPE_API}/${out_type}    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any

Create With A Duplicate Sequence Prefix Fails
    [Documentation]    INV-SEQ-001: the prefix is the middle segment of the transfer number, so
    ...    two types of one org sharing it could only be told apart by their counters.
    [Tags]    negative
    ${resp}=    GET On Session    api    ${STOCK_OPERATION_TYPE_API}/${STOCK_OPERATION_TYPE_ID}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}
    Response Status Should Be    ${resp}    200
    ${prefix}=    Set Variable    ${resp.json()}[item][sequence_prefix]
    ${name}=    Unique Display Name    Robot Duplicate Prefix
    ${resp}=    Post Stock Operation Type    ${{ {'name': {'en-US': $name}, 'operation_code': 'incoming', 'reservation_method': 'manual', 'backorder_policy': 'ask', 'sequence_prefix': $prefix, 'org_id': $INV_ORG_ID} }}
    Response Should Be Duplicate Values Error    ${resp}    sequence_prefix    org_id

Create With A Lowercase Sequence Prefix Fails
    [Documentation]    sequence_prefix is a resource_code: uppercase letters, digits and hyphens.
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Lowercase Prefix
    ${resp}=    Post Stock Operation Type    ${{ {'name': {'en-US': $name}, 'operation_code': 'incoming', 'reservation_method': 'manual', 'backorder_policy': 'ask', 'sequence_prefix': 'in/x', 'org_id': $INV_ORG_ID} }}
    Should Not Be Equal As Integers    ${resp.status_code}    201
    ...    msg=A sequence_prefix outside the resource_code form must not be accepted

Create With Default Locations Succeeds
    [Documentation]    The default source and destination are what a transfer of this type
    ...    pre-fills, so both must accept a real location reference.
    ${name}=    Unique Display Name    Robot Located Type
    ${resp}=    Post Stock Operation Type    ${{ {'name': {'en-US': $name}, 'operation_code': 'internal', 'reservation_method': 'manual', 'backorder_policy': 'never', 'default_source_location_id': $INVENTORY_LOCATION_ID, 'default_destination_location_id': $INVENTORY_LOCATION_ID, 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${resp}=    GET On Session    api    ${STOCK_OPERATION_TYPE_API}/${id}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/stock_operation_type.json    200
    Should Be Equal    ${item}[default_source_location_id]    ${INVENTORY_LOCATION_ID}
    DELETE On Session    api    ${STOCK_OPERATION_TYPE_API}/${id}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any

Create With Unknown Operation Code Fails
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Bad Operation
    ${resp}=    Post Stock Operation Type    ${{ {'name': {'en-US': $name}, 'operation_code': 'sideways', 'reservation_method': 'manual', 'backorder_policy': 'ask', 'org_id': $INV_ORG_ID} }}
    Should Not Be Equal As Integers    ${resp.status_code}    201
    ...    msg=An operation_code outside the declared enum must not be accepted

Create With Unknown Reservation Method Fails
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Bad Reservation
    ${resp}=    Post Stock Operation Type    ${{ {'name': {'en-US': $name}, 'operation_code': 'incoming', 'reservation_method': 'whenever', 'backorder_policy': 'ask', 'org_id': $INV_ORG_ID} }}
    Should Not Be Equal As Integers    ${resp.status_code}    201
    ...    msg=A reservation_method outside the declared enum must not be accepted

Create With Missing Required Fields Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${STOCK_OPERATION_TYPE_API}    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    name    operation_code    org_id
    ...    sequence_prefix    warehouse_id
