*** Settings ***
Documentation     Creating Product Categories. The first test saves the category under test
...               (${PRODUCT_CATEGORY_ID}/${PRODUCT_CATEGORY_ETAG}) consumed by the later
...               suites and deleted last by 08_delete.robot.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Inventory Org
Test Tags         inventory    product_category    create


*** Test Cases ***
Create With Required Fields Succeeds
    ${name}=    Unique Display Name    Robot Root Category
    ${code}=    Unique Code    cat
    ${resp}=    POST On Session    api    ${PRODUCT_CATEGORY_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${PRODUCT_CATEGORY_ID}    ${id}
    Set Global Variable    ${PRODUCT_CATEGORY_ETAG}    ${etag}
    Set Global Variable    ${PRODUCT_CATEGORY_CODE}    ${code}

Create Without Parent Succeeds
    [Documentation]    parent_category_id is nullable: a root category must be creatable
    ...    before any parent exists, or the tree could never be bootstrapped.
    ${name}=    Unique Display Name    Robot Bootstrap Category
    ${code}=    Unique Code    bootstrapcat
    ${resp}=    POST On Session    api    ${PRODUCT_CATEGORY_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${resp}=    GET On Session    api    ${PRODUCT_CATEGORY_API}/${id}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/product_category.json    200
    Should Be True    ${{ not $item.get('parent_category_id') }}
    ...    msg=A category created without a parent should have none
    DELETE On Session    api    ${PRODUCT_CATEGORY_API}/${id}    expected_status=any

Create With Parent Succeeds
    [Documentation]    Points a new category at a real parent, forming the two-level tree
    ...    03_update.robot uses to exercise the cycle rule.
    Ensure Product Category Under Test
    ${name}=    Unique Display Name    Robot Nested Category
    ${code}=    Unique Code    nestedcat
    ${resp}=    POST On Session    api    ${PRODUCT_CATEGORY_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'parent_category_id': $PRODUCT_CATEGORY_ID, 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    DELETE On Session    api    ${PRODUCT_CATEGORY_API}/${id}    expected_status=any

Create With Duplicate Code And Org Fails
    [Documentation]    BR §6.4.1: (code, org_id) is a composite unique — the same code is
    ...    reusable across orgs, but not twice within one.
    [Tags]    negative
    Ensure Product Category Under Test
    ${name}=    Unique Display Name    Robot Duplicate Category
    ${resp}=    POST On Session    api    ${PRODUCT_CATEGORY_API}
    ...    json=${{ {'code': $PRODUCT_CATEGORY_CODE, 'name': {'en-US': $name}, 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Response Should Be Duplicate Values Error    ${resp}    code    org_id

Create With Missing Required Fields Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_CATEGORY_API}    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    code    name    org_id

Create With Malformed Payload Fails
    [Tags]    negative
    ${headers}=    Create Dictionary    Content-Type=application/json
    ${resp}=    POST On Session    api    ${PRODUCT_CATEGORY_API}
    ...    data={ "name": {"en-US": "broken",    headers=${headers}    expected_status=any
    Response Should Be Malformed Payload Error    ${resp}

Create With Nonexist Field Fails
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Nonexist Field
    ${code}=    Unique Code    nonexist
    ${resp}=    POST On Session    api    ${PRODUCT_CATEGORY_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'org_id': $INV_ORG_ID, 'bla_bla_field': 'x'} }}
    ...    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field

Create With Not Found Parent Fails
    [Documentation]    parent_category_id is a real FK; the exact error contract for a
    ...    dangling reference is not pinned by a dedicated assertion keyword, so only the
    ...    non-success is asserted here.
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Orphan Category
    ${code}=    Unique Code    orphancat
    ${resp}=    POST On Session    api    ${PRODUCT_CATEGORY_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'parent_category_id': $NOT_FOUND_ID, 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    201
    ...    msg=A category referencing a nonexistent parent must not be creatable
