*** Settings ***
Documentation     Archiving the Product Type under test, rotating the saved etag. The type
...               is unarchived again so the later suites see it live.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Product Type Under Test
Test Tags         inventory    product_type    archive


*** Test Cases ***
Archive Succeeds
    ${resp}=    POST On Session    api    ${PRODUCT_TYPE_API}/${PRODUCT_TYPE_ID}/archived
    ...    json=${{ {'etag': $PRODUCT_TYPE_ETAG, 'is_archived': True} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PRODUCT_TYPE_ETAG}
    IF    $etag is not None    Set Global Variable    ${PRODUCT_TYPE_ETAG}    ${etag}

Archived Type Is Still Readable
    [Documentation]    Archiving is visibility, not deletion: an existing template still
    ...    points at this type, and its detail page must be able to name it.
    ${resp}=    GET On Session    api    ${PRODUCT_TYPE_API}/${PRODUCT_TYPE_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/product_type.json    200
    Should Be Equal    ${item}[is_archived]    ${True}

Unarchive Succeeds
    ${resp}=    POST On Session    api    ${PRODUCT_TYPE_API}/${PRODUCT_TYPE_ID}/archived
    ...    json=${{ {'etag': $PRODUCT_TYPE_ETAG, 'is_archived': False} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PRODUCT_TYPE_ETAG}
    IF    $etag is not None    Set Global Variable    ${PRODUCT_TYPE_ETAG}    ${etag}

Archive With Not Found Id Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_TYPE_API}/${NOT_FOUND_ID}/archived
    ...    json=${{ {'etag': $PRODUCT_TYPE_ETAG, 'is_archived': True} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Archive With Unmatched Etag Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_TYPE_API}/${PRODUCT_TYPE_ID}/archived
    ...    json=${{ {'etag': '___________________', 'is_archived': True} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Archive With Missing Required Fields Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_TYPE_API}/${PRODUCT_TYPE_ID}/archived
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag    is_archived
