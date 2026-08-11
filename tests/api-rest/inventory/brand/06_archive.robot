*** Settings ***
Documentation     Archiving the Brand under test, rotating the saved etag. The brand is
...               unarchived again so the later suites see it live.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Brand Under Test
Test Tags         inventory    brand    archive


*** Test Cases ***
Archive Succeeds
    ${resp}=    POST On Session    api    ${BRAND_API}/${BRAND_ID}/archived
    ...    json=${{ {'etag': $BRAND_ETAG, 'is_archived': True} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${BRAND_ETAG}
    IF    $etag is not None    Set Global Variable    ${BRAND_ETAG}    ${etag}

Archived Brand Is Still Readable
    [Documentation]    Archiving is visibility, not deletion: an existing template still
    ...    points at this brand, and its detail page must be able to name it.
    ${resp}=    GET On Session    api    ${BRAND_API}/${BRAND_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/brand.json    200
    Should Be Equal    ${item}[is_archived]    ${True}

Unarchive Succeeds
    ${resp}=    POST On Session    api    ${BRAND_API}/${BRAND_ID}/archived
    ...    json=${{ {'etag': $BRAND_ETAG, 'is_archived': False} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${BRAND_ETAG}
    IF    $etag is not None    Set Global Variable    ${BRAND_ETAG}    ${etag}

Archive With Not Found Id Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${BRAND_API}/${NOT_FOUND_ID}/archived
    ...    json=${{ {'etag': $BRAND_ETAG, 'is_archived': True} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Archive With Unmatched Etag Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${BRAND_API}/${BRAND_ID}/archived
    ...    json=${{ {'etag': '___________________', 'is_archived': True} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Archive With Missing Required Fields Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${BRAND_API}/${BRAND_ID}/archived
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag    is_archived
