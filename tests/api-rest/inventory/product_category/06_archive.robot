*** Settings ***
Documentation     Archiving the Product Category under test, rotating the saved etag. The
...               category is unarchived again so the later suites see it live.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Product Category Under Test
Test Tags         inventory    product_category    archive


*** Test Cases ***
Archive Succeeds
    ${resp}=    POST On Session    api    ${PRODUCT_CATEGORY_API}/${PRODUCT_CATEGORY_ID}/archived
    ...    json=${{ {'etag': $PRODUCT_CATEGORY_ETAG, 'is_archived': True} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PRODUCT_CATEGORY_ETAG}
    IF    $etag is not None    Set Global Variable    ${PRODUCT_CATEGORY_ETAG}    ${etag}

Archived Category Is Still Readable
    [Documentation]    Archiving is visibility, not deletion: a child category still points
    ...    at this one, and its detail page must be able to name it.
    ${resp}=    GET On Session    api    ${PRODUCT_CATEGORY_API}/${PRODUCT_CATEGORY_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/product_category.json    200
    Should Be Equal    ${item}[is_archived]    ${True}

Unarchive Succeeds
    ${resp}=    POST On Session    api    ${PRODUCT_CATEGORY_API}/${PRODUCT_CATEGORY_ID}/archived
    ...    json=${{ {'etag': $PRODUCT_CATEGORY_ETAG, 'is_archived': False} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PRODUCT_CATEGORY_ETAG}
    IF    $etag is not None    Set Global Variable    ${PRODUCT_CATEGORY_ETAG}    ${etag}

Archive With Not Found Id Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_CATEGORY_API}/${NOT_FOUND_ID}/archived
    ...    json=${{ {'etag': $PRODUCT_CATEGORY_ETAG, 'is_archived': True} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Archive With Unmatched Etag Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_CATEGORY_API}/${PRODUCT_CATEGORY_ID}/archived
    ...    json=${{ {'etag': '___________________', 'is_archived': True} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Archive With Missing Required Fields Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_CATEGORY_API}/${PRODUCT_CATEGORY_ID}/archived
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag    is_archived
