*** Settings ***
Documentation     Creating UoM Categories. The first test saves the category under test
...               (${UOMCAT_ID}/${UOMCAT_ETAG}) consumed by the later suites and deleted
...               last by 08_delete.robot.
Resource          resources/essential.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Org Under Test
Test Tags         essential    uomcat    create


*** Test Cases ***
Create With Required Fields Succeeds
    ${name}=    Unique Display Name    Robot Weight Category
    ${resp}=    POST On Session    api    ${UOMCAT_API}
    ...    json=${{ {'name': {'en-US': $name}, 'org_id': $UOM_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${UOMCAT_ID}    ${id}
    Set Global Variable    ${UOMCAT_ETAG}    ${etag}
    Set Global Variable    ${UOMCAT_NAME}    ${name}

Create Without Reference Uom Succeeds
    [Documentation]    BR-UOM-ESS-003: reference_uom_id is nullable on purpose. A category
    ...    must be creatable before the UoM that becomes its reference exists, or the two
    ...    resources could never be bootstrapped.
    ${name}=    Unique Display Name    Robot Bootstrap Category
    ${resp}=    POST On Session    api    ${UOMCAT_API}
    ...    json=${{ {'name': {'en-US': $name}, 'org_id': $UOM_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${resp}=    GET On Session    api    ${UOMCAT_API}/${id}
    ${item}=    Item Should Match Schema    ${resp}    ${ESSENTIAL_SCHEMA_DIR}/uomcat.json    200
    Should Be True    ${{ not $item.get('reference_uom_id') }}
    ...    msg=A category created without a reference UoM should have none
    DELETE On Session    api    ${UOMCAT_API}/${id}    expected_status=any

Create With Missing Required Fields Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${UOMCAT_API}    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    name    org_id

Create With Malformed Payload Fails
    [Tags]    negative
    ${headers}=    Create Dictionary    Content-Type=application/json
    ${resp}=    POST On Session    api    ${UOMCAT_API}
    ...    data={ "name": {"en-US": "broken",    headers=${headers}    expected_status=any
    Response Should Be Malformed Payload Error    ${resp}

Create With Nonexist Field Fails
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Nonexist Field
    ${resp}=    POST On Session    api    ${UOMCAT_API}
    ...    json=${{ {'name': {'en-US': $name}, 'org_id': $UOM_ORG_ID, 'bla_bla_field': 'x'} }}
    ...    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field

Create With Foreign Reference Uom Fails
    [Documentation]    BR-UOM-ESS-004 / UOM-ESS-INV-03: the Reference UoM must belong to
    ...    the category that names it. Pointing at a UoM from another category would make
    ...    every factor in this one meaningless.
    [Tags]    negative
    Ensure Foreign Uom Category
    ${name}=    Unique Display Name    Robot Foreign Reference
    ${resp}=    POST On Session    api    ${UOMCAT_API}
    ...    json=${{ {'name': {'en-US': $name}, 'org_id': $UOM_ORG_ID, 'reference_uom_id': $FOREIGN_UOM_ID} }}
    ...    expected_status=any
    Response Should Be Uomcat Foreign Reference Error    ${resp}

Create With Not Found Reference Uom Fails
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Missing Reference
    ${resp}=    POST On Session    api    ${UOMCAT_API}
    ...    json=${{ {'name': {'en-US': $name}, 'org_id': $UOM_ORG_ID, 'reference_uom_id': $NOT_FOUND_ID} }}
    ...    expected_status=any
    Response Should Be Uomcat Reference Not Found Error    ${resp}
