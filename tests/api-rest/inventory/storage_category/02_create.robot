*** Settings ***
Documentation     Creating Storage Categories, and the guard that stops one being archived
...               while a location still points at it.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Inventory Org
Test Tags         inventory    storage_category    create


*** Test Cases ***
Create With Required Fields Succeeds
    ${name}=    Unique Display Name    Robot Storage Category
    ${code}=    Unique Code    storcat
    ${resp}=    POST On Session    api    ${STORAGE_CATEGORY_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'allow_new_item_policy': 'allow', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${STORAGE_CATEGORY_ID}    ${id}
    Set Global Variable    ${STORAGE_CATEGORY_ETAG}    ${etag}
    Set Global Variable    ${STORAGE_CATEGORY_CODE}    ${code}

Create With A Weight Limit Succeeds
    ${name}=    Unique Display Name    Robot Pallet Zone
    ${code}=    Unique Code    pallet
    ${resp}=    POST On Session    api    ${STORAGE_CATEGORY_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'max_weight': '1000', 'allow_new_item_policy': 'empty_only', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${resp}=    GET On Session    api    ${STORAGE_CATEGORY_API}/${id}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/storage_category.json    200
    Should Be Equal    ${item}[allow_new_item_policy]    empty_only
    DELETE On Session    api    ${STORAGE_CATEGORY_API}/${id}    expected_status=any

Create With Duplicate Code Fails
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Duplicate Category
    ${resp}=    POST On Session    api    ${STORAGE_CATEGORY_API}
    ...    json=${{ {'code': $STORAGE_CATEGORY_CODE, 'name': {'en-US': $name}, 'allow_new_item_policy': 'allow', 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Response Should Be Duplicate Values Error    ${resp}

Archiving A Category A Location Uses Is Refused
    [Documentation]    Withdrawing a category while a live location cites it would leave that
    ...    location naming a policy nobody can look up. The location is reassigned first.
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Categorised Location
    ${code}=    Unique Code    catloc
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'location_usage': 'internal', 'storage_category_id': $STORAGE_CATEGORY_ID, 'org_id': $INV_ORG_ID} }}
    ${location_id}    ${location_etag}=    Response Should Be Create Success    ${resp}

    ${resp}=    GET On Session    api    ${STORAGE_CATEGORY_API}/${STORAGE_CATEGORY_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/storage_category.json    200
    ${resp}=    POST On Session    api    ${STORAGE_CATEGORY_API}/${STORAGE_CATEGORY_ID}/archived
    ...    json=${{ {'is_archived': True, 'etag': $item['etag']} }}    expected_status=any
    Should Be True    ${resp.status_code} >= 400
    ...    msg=A category a live location uses must not archive

    DELETE On Session    api    ${INVENTORY_LOCATION_API}/${location_id}    expected_status=any

Archiving An Unused Category Succeeds
    [Documentation]    With nothing pointing at it, the category leaves the working set
    ...    cleanly. It is unarchived again so the later suites still have it.
    ${resp}=    GET On Session    api    ${STORAGE_CATEGORY_API}/${STORAGE_CATEGORY_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/storage_category.json    200
    ${resp}=    POST On Session    api    ${STORAGE_CATEGORY_API}/${STORAGE_CATEGORY_ID}/archived
    ...    json=${{ {'is_archived': True, 'etag': $item['etag']} }}
    Response Status Should Be    ${resp}    200

    ${resp}=    GET On Session    api    ${STORAGE_CATEGORY_API}/${STORAGE_CATEGORY_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/storage_category.json    200
    Should Be True    ${item}[is_archived]
    ${resp}=    POST On Session    api    ${STORAGE_CATEGORY_API}/${STORAGE_CATEGORY_ID}/archived
    ...    json=${{ {'is_archived': False, 'etag': $item['etag']} }}
    Response Status Should Be    ${resp}    200
