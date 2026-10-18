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

Create Without Parent Lands Under All
    [Documentation]    Every category has a parent. One created without lands under the org's
    ...    root "All", and its canonical name starts from the root's.
    Ensure Root Product Category
    ${name}=    Unique Display Name    Robot Bootstrap Category
    ${code}=    Unique Code    bootstrapcat
    ${resp}=    POST On Session    api    ${PRODUCT_CATEGORY_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name, 'vi-VN': $name + ' VI'}, 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${item}=    Get Product Category    ${id}
    Should Be Equal    ${item}[parent_category_id]    ${ROOT_CATEGORY_ID}
    Should Be Equal    ${item}[canonical_name][en-US]    ${ROOT_CATEGORY_CANONICAL}[en-US] / ${name}
    Should Be Equal    ${item}[canonical_name][vi-VN]    ${ROOT_CATEGORY_CANONICAL}[vi-VN] / ${name} VI
    Should Be True    ${{ $item.get('is_root') is None }}    msg=Only the root carries is_root
    [Teardown]    DELETE On Session    api    ${PRODUCT_CATEGORY_API}/${id}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any

Create With Parent Succeeds
    [Documentation]    Points a new category at a real parent, forming the two-level tree
    ...    03_update.robot uses to exercise the cycle rule. The child has no vi-VN name while
    ...    the chain does, so its vi-VN segment falls back to its en-US name.
    Ensure Product Category Under Test
    ${parent}=    Get Product Category    ${PRODUCT_CATEGORY_ID}
    ${name}=    Unique Display Name    Robot Nested Category
    ${code}=    Unique Code    nestedcat
    ${resp}=    POST On Session    api    ${PRODUCT_CATEGORY_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'parent_category_id': $PRODUCT_CATEGORY_ID, 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${item}=    Get Product Category    ${id}
    Should Be Equal    ${item}[canonical_name][en-US]    ${parent}[canonical_name][en-US] / ${name}
    Should Be Equal    ${item}[canonical_name][vi-VN]    ${parent}[canonical_name][vi-VN] / ${name}
    [Teardown]    DELETE On Session    api    ${PRODUCT_CATEGORY_API}/${id}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any

Search By Canonical Name Succeeds
    [Documentation]    canonical_name is a persisted column, so it is filterable like any other.
    Ensure Product Category Under Test
    ${category}=    Get Product Category    ${PRODUCT_CATEGORY_ID}
    ${graph}=    Evaluate    json.dumps({"if": ["canonical_name", "=", $category['canonical_name']['en-US']]})    json
    ${resp}=    GET On Session    api    ${PRODUCT_CATEGORY_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'graph': $graph, 'language': 'en-US'} }}
    Response Status Should Be    ${resp}    200
    Search Results Should Contain Id    ${resp}    ${PRODUCT_CATEGORY_ID}

Create With Canonical Name Fails
    [Documentation]    canonical_name is derived from the tree; a client may never set it.
    [Tags]    negative
    ${code}=    Unique Code    canoncat
    ${resp}=    POST On Session    api    ${PRODUCT_CATEGORY_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': 'X'}, 'canonical_name': {'en-US': 'Fake / X'}, 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Response Should Be Category Server Managed Error    ${resp}    canonical_name

Create With Is Root Fails
    [Documentation]    The root is seeded; the API never creates another one.
    [Tags]    negative
    ${code}=    Unique Code    rootcat
    ${resp}=    POST On Session    api    ${PRODUCT_CATEGORY_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': 'X'}, 'is_root': True, 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Response Should Be Category Server Managed Error    ${resp}    is_root

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
    [Documentation]    Without an org there is no root to default the parent to, so the
    ...    parent is reported missing alongside the other required fields.
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_CATEGORY_API}    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    code    name    parent_category_id    org_id

Create With Malformed Payload Fails
    [Tags]    negative
    ${headers}=    Create Dictionary    Content-Type=application/json
    ${resp}=    POST On Session    api    ${PRODUCT_CATEGORY_API}
    ...    data={ "name": {"en-US": "broken",    headers=${headers}
    ...    json=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
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
