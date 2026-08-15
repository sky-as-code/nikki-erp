*** Settings ***
Documentation     Creating Inventory Locations. The first test saves the location under test
...               (${INVENTORY_LOCATION_ID}/${INVENTORY_LOCATION_ETAG}) consumed by the later
...               suites and deleted last by 08_delete.robot.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Inventory Org
Test Tags         inventory    inventory_location    create


*** Test Cases ***
Create With Required Fields Succeeds
    ${name}=    Unique Display Name    Robot Inventory Location
    ${code}=    Unique Code    stkloc
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'location_usage': 'internal', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${INVENTORY_LOCATION_ID}    ${id}
    Set Global Variable    ${INVENTORY_LOCATION_ETAG}    ${etag}
    Set Global Variable    ${INVENTORY_LOCATION_CODE}    ${code}

Create With Parent Location Succeeds
    [Documentation]    Locations form a tree, so a child may point at the location under test.
    ${name}=    Unique Display Name    Robot Child Location
    ${code}=    Unique Code    childloc
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'location_usage': 'internal', 'parent_location_id': $INVENTORY_LOCATION_ID, 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${resp}=    GET On Session    api    ${INVENTORY_LOCATION_API}/${id}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/inventory_location.json    200
    Should Be Equal    ${item}[parent_location_id]    ${INVENTORY_LOCATION_ID}
    Set Global Variable    ${CHILD_INVENTORY_LOCATION_ID}    ${id}

Create Each Virtual Location Type Succeeds
    [Documentation]    BR §4.2.7 and §4.2.9: an adjustment balances against an
    ...    inventory_loss location and a scrap moves to a scrap location, so both must be
    ...    creatable before those flows can exist at all.
    FOR    ${type}    IN    customer    vendor    inventory_loss    scrap    transit
        ${name}=    Unique Display Name    Robot ${type} Location
        ${code}=    Unique Code    ${type}
        ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}
        ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'location_usage': $type, 'org_id': $INV_ORG_ID} }}
        ${id}    ${etag}=    Response Should Be Create Success    ${resp}
        DELETE On Session    api    ${INVENTORY_LOCATION_API}/${id}    expected_status=any
    END

Create Defaults Location Type To Internal
    [Documentation]    The schema defaults location_usage to `internal`, the only type that
    ...    holds company-owned stock — the safe default for a location created without one.
    ${name}=    Unique Display Name    Robot Default Type Location
    ${code}=    Unique Code    defloc
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${resp}=    GET On Session    api    ${INVENTORY_LOCATION_API}/${id}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/inventory_location.json    200
    Should Be Equal    ${item}[location_usage]    internal
    DELETE On Session    api    ${INVENTORY_LOCATION_API}/${id}    expected_status=any

Create With Unknown Location Type Fails
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Bad Type Location
    ${code}=    Unique Code    badtype
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'location_usage': 'warehouse', 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    201
    ...    msg=A location_usage outside the declared enum must not be accepted

Create With Duplicate Code And Org Fails
    [Documentation]    composite_uniques on inventory_location.json is ["code", "org_id"]: the
    ...    same code is only ambiguous within one org.
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Duplicate Location
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}
    ...    json=${{ {'code': $INVENTORY_LOCATION_CODE, 'name': {'en-US': $name}, 'location_usage': 'internal', 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Response Should Be Duplicate Values Error    ${resp}    code    org_id

Create With Missing Required Fields Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    code    name    org_id

Create With Archived Flag Fails
    [Documentation]    is_archived is set through the /archived action, never at create: a
    ...    caller creating an already-invisible record is telling us the request is wrong.
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Archived Location
    ${code}=    Unique Code    archloc
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'location_usage': 'internal', 'org_id': $INV_ORG_ID, 'is_archived': True} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    201
    ...    msg=is_archived must not be accepted at create time
